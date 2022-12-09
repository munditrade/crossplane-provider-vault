package vault

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/munditrade/provider-secret/internal/common"
	"log"
)

func New(props map[string][]byte) (common.SecretManager, error) {
	host := string(props["host"])
	port := string(props["port"])
	token := string(props["token"])

	return NewVaultSecretManager(host, port, token)
}

type VaultSecretManager struct {
	client *vault.Client
}

func NewVaultSecretManager(host, port, token string) (*VaultSecretManager, error) {
	config := vault.DefaultConfig()

	config.Address = fmt.Sprintf("%s:%s", host, port)

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	client.SetToken(token)

	if err != nil {
		return nil, fmt.Errorf("vault login error: %w", err)
	}

	return &VaultSecretManager{client: client}, nil
}

func (m *VaultSecretManager) Put(ctx context.Context, engine string, secretPath string, data map[string]interface{}) error {
	_, err := m.client.KVv2(engine).Put(ctx, secretPath, data)
	return err
}

func (m *VaultSecretManager) CreateEngine(ctx context.Context, engine string, engineType string, options map[string]string) error {
	return m.client.Sys().MountWithContext(ctx, engine, &vault.MountInput{
		Type:    engineType,
		Options: options,
	})
}

func (m *VaultSecretManager) ExistEngine(ctx context.Context, engine string) (bool, error) {
	k, _ := m.client.Sys().MountConfigWithContext(ctx, engine)

	return k != nil, nil
}

func (m *VaultSecretManager) DeleteEngine(ctx context.Context, engine string) error {
	return m.client.Sys().Unmount(engine)
}
