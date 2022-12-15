# provider-secret

Prepare Env

```
$ make submodules
```

# Installation

```
$ kubectl create secret generic secret-conn \
  --from-literal=host=http://127.0.0.1 \
  --from-literal=port=8200 \
  --from-literal=token={YOUR_VAULT_TOKEN}
```

# Vault 

1. Install vault

```
$ helm repo add hashicorp https://helm.releases.hashicorp.com
$ helm install vault hashicorp/vault
```