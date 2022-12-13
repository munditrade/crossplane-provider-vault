package controller

import (
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	vaultV1alpha "github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/munditrade/provider-secret/internal/clients/vault"
	"github.com/munditrade/provider-secret/internal/controller/engine"
	"github.com/munditrade/provider-secret/internal/controller/secretpath"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/munditrade/provider-secret/internal/controller/config"
)

// Setup creates all Template controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		engine.Setup(vault.New, &vaultV1alpha.Engine{}),
		secretpath.Setup(vault.New, &vaultV1alpha.SecretPath{}),
		config.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
