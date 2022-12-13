package common

import "context"

type GetNewSecretManager func(props map[string][]byte) (SecretManager, error)

const (
	ErrNotFoundPath = "path does not exist"
)

type SecretManager interface {
	Put(ctx context.Context, engine string, secretPath string, data map[string]interface{}, options map[string]string) error
	GetSecrets(ctx context.Context, engine string, secretPath string, options map[string]string) (map[string]interface{}, error)
	CreateEngine(ctx context.Context, engine string, engineType string, options map[string]string) error
	ExistEngine(ctx context.Context, engine string) (bool, error)
	DeletePath(ctx context.Context, engine string, secretPath string, options map[string]string) error
	DeleteEngine(ctx context.Context, engine string) error
}
