# provider-secret

Prepare Env

```
$ make submodules
```

# Installation

```
$ kubectl create secret generic secret-conn \
  --from-literal=host=http://{vault_host} \
  --from-literal=port={vault_port} \ // Default 8200
  --from-literal=token={YOUR_VAULT_TOKEN}
```

# Vault 

1. Install vault

```
$ helm repo add hashicorp https://helm.releases.hashicorp.com
$ helm install vault hashicorp/vault
```
