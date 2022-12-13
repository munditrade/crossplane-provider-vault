package common

import (
	"context"
	"github.com/munditrade/provider-secret/apis/vault/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ErrNoParentReferences = "CR does not have parent ref"
)

func getOwnerEngine(ctx context.Context, reader client.Reader, ns string, engineName string) (*v1alpha1.Engine, error) {
	engine := new(v1alpha1.Engine)
	err := reader.Get(ctx, types.NamespacedName{Namespace: ns, Name: engineName}, engine)
	if err == nil && engine.ObjectMeta.Name == engineName {
		return engine, nil
	}

	return nil, errors.New(ErrNoParentReferences)
}

func GetOwnerEngine(ctx context.Context, reader client.Reader, ns string, engineName string) (*v1alpha1.Engine, error) {
	return getOwnerEngine(ctx, reader, ns, engineName)
}
