package vault

import (
	"context"
	"fmt"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/namespace"
	"github.com/hashicorp/vault/vault"
	"github.com/munditrade/provider-secret/internal/clients"
	"github.com/munditrade/provider-secret/internal/clients/exceptions"
	"log"
	"strings"
)

func NewVaultPolicyManager(props map[string][]byte) (clients.PolicyManager, error) {
	host := string(props["host"])
	port := string(props["port"])
	token := string(props["token"])

	return NewPolicyManager(host, port, token)
}

type PolicyManager struct {
	client *vaultApi.Client
}

func NewPolicyManager(host string, port string, token string) (*PolicyManager, error) {
	config := vaultApi.DefaultConfig()

	config.Address = fmt.Sprintf("%s:%s", host, port)

	client, err := vaultApi.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	client.SetToken(token)

	if err != nil {
		return nil, fmt.Errorf("vault login error: %w", err)
	}

	return &PolicyManager{client: client}, nil
}

func (p *PolicyManager) Put(ctx context.Context, name string, policy clients.Policy) error {
	rules := strings.Trim(fmt.Sprintf(`path "%s" { capabilities = ["%s"] }`,
		policy.PathConfig.Path,
		strings.Join(policy.PathConfig.Capabilities, "\",\""),
	), "")

	return p.client.Sys().PutPolicyWithContext(ctx, name, rules)
}

func (p *PolicyManager) Delete(ctx context.Context, name string) error {
	return p.client.Sys().DeletePolicyWithContext(ctx, name)
}

func (p *PolicyManager) Get(ctx context.Context, name string) ([]clients.Policy, error) {
	policyAsStr, err := p.client.Sys().GetPolicyWithContext(ctx, name)

	if err == nil && policyAsStr == "" {
		return nil, exceptions.NewNotFoundPolicy(name)
	}

	if err != nil {
		return nil, err
	}

	policies := make([]clients.Policy, 0)
	aclPolicy, parseErr := vault.ParseACLPolicy(namespace.RootNamespace, policyAsStr)

	if parseErr != nil {
		return nil, parseErr
	}

	for _, path := range aclPolicy.Paths {
		policies = append(policies, clients.Policy{PathConfig: clients.PathConfig{
			Path:         path.Path,
			Capabilities: path.Capabilities,
		}})
	}

	return policies, nil
}
