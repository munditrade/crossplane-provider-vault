package secretpath

import (
	"context"
	"fmt"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/munditrade/provider-secret/internal/common"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha12 "github.com/munditrade/provider-secret/apis/secret/v1alpha1"
	"github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/munditrade/provider-secret/internal/controller/features"
)

const (
	errNotSecretPath      = "managed resource is not a SecretPath custom resource"
	errTrackPCUsage       = "cannot track ProviderConfig usage"
	errGetPC              = "cannot get ProviderConfig"
	errGetCreds           = "cannot get credentials"
	errEngineNotFound     = "engine not found"
	errDataNotFoundInPath = "cannot get data from path"
	errCreatingPath       = "error during path creation"
	errNoSecretRef        = "ProviderConfig does not reference a credentials Secret"
	errNewClient          = "cannot create new Service"
)

var emptyMap = map[string]interface{}{
	"secret": "empty",
}

// Setup adds a controller that reconciles SecretPath managed resources.
func Setup(getNewSecretManager common.GetNewSecretManager, object client.Object) func(mgr ctrl.Manager, o controller.Options) error {
	return func(mgr ctrl.Manager, o controller.Options) error {
		name := managed.ControllerName(v1alpha1.SecretPathGroupKind)

		cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
		if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
			cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha12.StoreConfigGroupVersionKind))
		}

		r := managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SecretPathGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kubeAPI:          mgr.GetClient(),
				usage:            resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1alpha12.ProviderConfigUsage{}),
				newSecretManager: getNewSecretManager}),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...))

		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			WithOptions(o.ForControllerRuntime()).
			For(object).
			Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
	}
}

type connector struct {
	kubeAPI          common.K8sReader
	usage            resource.Tracker
	newSecretManager func(props map[string][]byte) (common.SecretManager, error)
}

type external struct {
	kubeReader common.K8sReader
	service    common.SecretManager
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.SecretPath)
	if !ok {
		return nil, errors.New(errNotSecretPath)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &v1alpha12.ProviderConfig{}
	if err := c.kubeAPI.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	ref := pc.Spec.Credentials.ConnectionSecretRef
	if ref == nil {
		return nil, errors.New(errNoSecretRef)
	}

	s := &corev1.Secret{}
	if err := c.kubeAPI.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s); err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newSecretManager(s.Data)

	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc, kubeReader: c.kubeAPI}, nil
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.SecretPath)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSecretPath)
	}

	engine, err := common.GetOwnerEngine(ctx, c.kubeReader, cr.Namespace, cr.Spec.ForProvider.Engine)

	if engine == nil {
		return managed.ExternalObservation{}, errors.New(errEngineNotFound)
	}

	if err != nil {
		return managed.ExternalObservation{}, err
	}

	storage := engine.ObjectMeta.Name
	path := cr.Spec.ForProvider.Path

	_, getSecretsErr := c.service.GetSecrets(ctx, storage, path, engine.Spec.ForProvider.Options)

	if getSecretsErr != nil {
		if getSecretsErr.Error() == common.ErrNotFoundPath {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}

		return managed.ExternalObservation{}, errors.New(errDataNotFoundInPath)
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SecretPath)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSecretPath)
	}

	engine, err := common.GetOwnerEngine(ctx, c.kubeReader, cr.Namespace, cr.Spec.ForProvider.Engine)

	if engine == nil {
		return managed.ExternalCreation{}, errors.New(errEngineNotFound)
	}

	if err != nil {
		return managed.ExternalCreation{}, err
	}

	storage := engine.ObjectMeta.Name
	path := cr.Spec.ForProvider.Path
	engineOpts := engine.Spec.ForProvider.Options

	if putErr := c.service.Put(ctx, storage, path, emptyMap, engineOpts); putErr != nil {
		return managed.ExternalCreation{}, errors.New(errCreatingPath)
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.SecretPath)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSecretPath)
	}

	fmt.Printf("Updating: %+v", cr)

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.SecretPath)
	if !ok {
		return errors.New(errNotSecretPath)
	}

	engine, err := common.GetOwnerEngine(ctx, c.kubeReader, cr.Namespace, cr.Spec.ForProvider.Engine)

	if err != nil {
		if err.Error() == common.ErrNoParentReferences {
			return nil
		}

		return err
	}

	storage := engine.ObjectMeta.Name
	path := cr.Spec.ForProvider.Path
	engineOpts := engine.Spec.ForProvider.Options

	if err := c.service.DeletePath(ctx, storage, path, engineOpts); err != nil {
		if err.Error() == common.ErrNotFoundPath {
			return nil
		}

		return err
	}

	return nil
}
