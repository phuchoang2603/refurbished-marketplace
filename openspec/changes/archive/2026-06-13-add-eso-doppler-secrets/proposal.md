## Why

GHCR images are ready (#6) and ArgoCD is next (#4), but secrets still lived in Git as plaintext dev credentials (`infra/k8s/secrets.yaml`). That blocked safe non-local deploys and duplicated secret paths between local Tilt and future GitOps clusters. External Secrets Operator with Doppler is the first provider — same mechanism for local `dev` and remote `prd` configs.

## What Changes

- Install External Secrets Operator (upstream Helm) in the `operators` namespace for local Tilt
- Add cluster-level ESO manifests under `infra/k8s/` (`ClusterSecretStore` + gitignored `doppler-token.secret.yaml` from devenv)
- Render application `ExternalSecret` resources from the `refurbished-marketplace` Helm chart, derived from `services.<slug>.db` and `services.<slug>.auth`
- **BREAKING:** Remove `infra/k8s/secrets.yaml`; delete Tilt `k8s_yaml` for committed secrets
- Bootstrap Doppler **service token** (`dp.st…`) via `.env` + devenv `files` → `infra/k8s/doppler-token.secret.yaml` (`dopplerToken` key)
- Set `DOPPLER_PROJECT` / `DOPPLER_CONFIG` in `devenv.nix`; document service token setup in `docs/development/secrets.md`
- Sync secrets into existing K8s Secret names (`users-app`, `products-app`, `orders-app`, `payment-app`, `users-auth`) so deployment templates stay unchanged
- Document provider swap (Doppler → AWS Secrets Manager, Vault, etc.) via `ClusterSecretStore` and marketplace chart values

### Non-goals

- ArgoCD bootstrap (#4) — prerequisite, not included here
- GHCR `imagePullSecrets` via ESO — follow-up
- Doppler OIDC / workload identity — service tokens first
- Dedicated Helm chart only for ESO — app ExternalSecrets live in the marketplace chart
- Preflight `secrets-check` scripts — ESO CR status is sufficient

## Capabilities

### New Capabilities

- `external-secrets`: ESO + Doppler secret sync for local and remote Kubernetes, replacing committed plaintext secrets

### Modified Capabilities

<!-- none — service deployment templates unchanged; only secret provisioning changes -->

## Impact

- **Remove:** `infra/k8s/secrets.yaml`
- **Add:** `infra/k8s/cluster-secret-store.yaml`, devenv-generated `infra/k8s/doppler-token.secret.yaml`, `templates/external-secrets.tpl` in marketplace chart
- **Update:** `Tiltfile`, `devenv.nix`, `infra/charts/refurbished-marketplace/values.yaml`, development docs
- **Issue #10:** implementation tracking; blocks #4 ArgoCD for non-local deploy
- **Unchanged:** service deployment and `kafka` chart templates, application env/secretKeyRef usage
