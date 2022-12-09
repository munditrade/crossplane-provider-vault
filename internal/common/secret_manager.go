package common

import "context"

type GetNewSecretManager func(props map[string][]byte) (SecretManager, error)

type SecretManager interface {
	Put(ctx context.Context, engine string, secretPath string, data map[string]interface{}) error
	CreateEngine(ctx context.Context, engine string, engineType string, options map[string]string) error
	ExistEngine(ctx context.Context, engine string) (bool, error)
	DeleteEngine(ctx context.Context, engine string) error
}
