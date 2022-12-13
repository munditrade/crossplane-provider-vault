package engine

import (
	"context"
	"fmt"
	v1alpha12 "github.com/munditrade/provider-secret/apis/secret/v1alpha1"
	"github.com/munditrade/provider-secret/internal/common"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	errNotEngine          = "managed resource is not a Engine custom resource"
	errCreatingEngine     = "creation engine error"
	errTrackPCUsage       = "cannot track ProviderConfig usage"
	errGetPC              = "cannot get ProviderConfig"
	errErrorGettingEngine = "cannot get path given a engine"
	errNoSecretRef        = "ProviderConfig does not reference a credentials Secret"
	errGetSecret          = "cannot get credentials Secret"

	errNewClient = "cannot create new Service"
)

// Setup adds a controller that reconciles Engine managed resources.
func Setup(getNewSecretManager common.GetNewSecretManager, object client.Object) func(mgr ctrl.Manager, o controller.Options) error {
	return func(mgr ctrl.Manager, o controller.Options) error {
		name := managed.ControllerName(v1alpha1.EngineGroupKind)

		cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
		if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
			cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha12.StoreConfigGroupVersionKind))
		}

		r := managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.EngineGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube:             mgr.GetClient(),
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

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube             client.Client
	usage            resource.Tracker
	newSecretManager func(props map[string][]byte) (common.SecretManager, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Engine)
	if !ok {
		return nil, errors.New(errNotEngine)
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

	svc, err := c.newSecretManager(s.Data)

	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

type external struct {
	service common.SecretManager
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Engine)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEngine)
	}

	engine := cr.ObjectMeta.Name

	exist, err := c.service.ExistEngine(ctx, engine)

	if err != nil {
		return managed.ExternalObservation{}, errors.New(errErrorGettingEngine)
	}

	// These fmt statements should be removed in the real implementation.
	fmt.Printf("Observing: %+v", cr)

	return managed.ExternalObservation{
		ResourceExists:    exist,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Engine)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEngine)
	}

	fmt.Printf("Creating: %+v", cr)

	engine := cr.ObjectMeta.Name
	storage := cr.Spec.ForProvider.Storage
	opts := cr.Spec.ForProvider.Options

	err := c.service.CreateEngine(ctx, engine, storage, opts)

	if err != nil {
		return managed.ExternalCreation{}, errors.New(errCreatingEngine)
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Engine)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEngine)
	}

	fmt.Printf("Updating: %+v", cr)

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Engine)
	if !ok {
		return errors.New(errNotEngine)
	}

	engine := cr.ObjectMeta.Name

	exist, _ := c.service.ExistEngine(ctx, engine)

	if exist {
		fmt.Printf("Deleting: %+v", cr)
		return c.service.DeleteEngine(ctx, engine)
	}

	return nil
}
