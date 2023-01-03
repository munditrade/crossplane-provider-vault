package policy

import (
	"context"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	v1alpha12 "github.com/munditrade/provider-secret/apis/secret/v1alpha1"
	"github.com/munditrade/provider-secret/internal/clients"
	"github.com/munditrade/provider-secret/internal/clients/exceptions"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/munditrade/provider-secret/internal/controller/features"
)

const (
	errNotPolicy    = "managed resource is not a Policy custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errNoSecretRef  = "ProviderConfig does not reference a credentials Secret"
	errGetSecret    = "cannot get credentials Secret"

	errNewClient = "cannot create new Service"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func policyChange(old clients.Policy, current clients.Policy) bool {
	currentCapabilities := current.PathConfig.Capabilities
	oldCapabilities := old.PathConfig.Capabilities

	oldPath := strings.Split(old.PathConfig.Path, "*")[0]
	newPath := strings.Split(current.PathConfig.Path, "*")[0]

	if oldPath != newPath {
		return true
	}

	if len(oldCapabilities) != len(currentCapabilities) {
		return true
	}

	for _, cap := range oldCapabilities {
		if !contains(currentCapabilities, cap) {
			return true
		}
	}

	return false
}

func getPolicy(rule v1alpha1.Rule) clients.Policy {
	path := rule.Path
	capabilities := rule.Capabilities

	return clients.Policy{
		PathConfig: clients.PathConfig{
			Path:         path,
			Capabilities: capabilities,
		},
	}
}

func Setup(newPolicyManager clients.GetPolicyManager) func(mgr ctrl.Manager, o controller.Options) error {
	return func(mgr ctrl.Manager, o controller.Options) error {
		name := managed.ControllerName(v1alpha1.PolicyGroupKind)

		cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
		if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
			cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha12.StoreConfigGroupVersionKind))
		}

		r := managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.PolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube:         mgr.GetClient(),
				usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1alpha12.ProviderConfigUsage{}),
				newServiceFn: newPolicyManager}),
			managed.WithLogger(o.Logger.WithValues("controller-policy", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...))

		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			WithOptions(o.ForControllerRuntime()).
			For(&v1alpha1.Policy{}).
			Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
	}

}

type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn clients.GetPolicyManager
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return nil, errors.New(errNotPolicy)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &v1alpha12.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	ref := pc.Spec.Credentials.ConnectionSecretRef
	if ref == nil {
		return nil, errors.New(errNoSecretRef)
	}

	s := &corev1.Secret{}

	if err := c.kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s); err != nil {
		return nil, errors.Wrap(err, errGetSecret)
	}

	svc, err := c.newServiceFn(s.Data)

	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

type external struct {
	service clients.PolicyManager
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	var policyNotFoundErr *exceptions.NotFoundPolicy

	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPolicy)
	}

	policyName := cr.ObjectMeta.Name
	oldPolicies, err := c.service.Get(ctx, policyName)

	if errors.As(err, &policyNotFoundErr) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "error getting policy")
	}

	oldPoliciesAsMap := make(map[string]clients.Policy)
	currentPolicyAsMap := make(map[string]clients.Policy)

	for _, rule := range cr.Spec.ForProvider.Rules {
		path := strings.Split(rule.Path, "*")[0]
		currentPolicyAsMap[path] = getPolicy(rule)
	}

	for _, policy := range oldPolicies {
		path := strings.Split(policy.PathConfig.Path, "*")[0]
		oldPoliciesAsMap[path] = policy
	}

	if len(oldPoliciesAsMap) != len(currentPolicyAsMap) {
		return managed.ExternalObservation{
			ResourceExists:    true,
			ResourceUpToDate:  false,
			ConnectionDetails: managed.ConnectionDetails{},
		}, nil
	}

	for oldPath, oldPolicy := range oldPoliciesAsMap {
		if newPolicy, exist := currentPolicyAsMap[oldPath]; exist {
			if policyChange(oldPolicy, newPolicy) {
				return managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: managed.ConnectionDetails{},
				}, nil
			}
		} else {
			return managed.ExternalObservation{
				ResourceExists:    true,
				ResourceUpToDate:  false,
				ConnectionDetails: managed.ConnectionDetails{},
			}, nil
		}
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPolicy)
	}

	policyName := cr.ObjectMeta.Name

	for _, rule := range cr.Spec.ForProvider.Rules {
		path := rule.Path
		capabilities := rule.Capabilities

		policy := clients.Policy{
			PathConfig: clients.PathConfig{
				Path:         path,
				Capabilities: capabilities,
			},
		}

		err := c.service.Put(ctx, policyName, policy)

		if err != nil {
			return managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{}}, err
		}
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPolicy)
	}

	policyName := cr.ObjectMeta.Name

	for _, rule := range cr.Spec.ForProvider.Rules {
		path := rule.Path
		capabilities := rule.Capabilities

		policy := clients.Policy{
			PathConfig: clients.PathConfig{
				Path:         path,
				Capabilities: capabilities,
			},
		}

		err := c.service.Put(ctx, policyName, policy)

		if err != nil {
			return managed.ExternalUpdate{ConnectionDetails: managed.ConnectionDetails{}}, err
		}
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return errors.New(errNotPolicy)
	}

	err := c.service.Delete(ctx, cr.ObjectMeta.Name)

	if err != nil {
		return err
	}

	return nil
}
