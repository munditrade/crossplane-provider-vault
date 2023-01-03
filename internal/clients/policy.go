package clients

import "context"

type GetPolicyManager func(props map[string][]byte) (PolicyManager, error)

type PolicyManager interface {
	Put(ctx context.Context, name string, policy Policy) error
	Delete(ctx context.Context, name string) error
	Get(ctx context.Context, name string) ([]Policy, error)
}

type Policy struct {
	PathConfig PathConfig
}

type PathConfig struct {
	Path         string
	Capabilities []string
}
