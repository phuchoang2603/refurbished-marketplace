## Why

GHCR images are ready (#6) and ArgoCD is next (#4), but secrets still live in Git as plaintext dev credentials (`infra/k8s/secrets.yaml`). That blocks safe non-local deploys and duplicates secret paths between local Tilt and future GitOps clusters. We need External Secrets Operator with Doppler as the first provider — same mechanism for local `dev` and remote `prd` configs.

## What Changes

- Install External Secrets Operator (upstream Helm) for local Tilt and document the same path for remote clusters
- Add flat ESO manifests under `infra/eso/` (`ClusterSecretStore` + `ExternalSecret` resources) — no dedicated Helm chart
- **BREAKING:** Remove `infra/k8s/secrets.yaml`; delete Tilt `k8s_yaml` for committed secrets
- Bootstrap Doppler **service token** (`dp.st…`) via `.env` + devenv dotenv → Tilt `local_resource` creates K8s Secret for ESO (`dopplerToken` key)
- Add Doppler CLI to `devenv.nix`; document `doppler login`, `doppler setup`, and service token creation
- Sync secrets into existing K8s Secret names (`users-app`, `products-app`, `orders-app`, `payment-app`, `users-auth`) so Helm charts stay unchanged
- Document provider swap (Doppler → AWS Secrets Manager, Vault, etc.) by changing `ClusterSecretStore` only

### Non-goals

- ArgoCD bootstrap (#4) — this change is a prerequisite, not included here
- GHCR `imagePullSecrets` via ESO — follow-up
- Doppler OIDC / workload identity — service tokens first
- Custom Helm chart wrapping ESO manifests
- Preflight `secrets-check` scripts — ESO CR status is sufficient

## Capabilities

### New Capabilities

- `external-secrets`: ESO + Doppler secret sync for local and remote Kubernetes, replacing committed plaintext secrets

### Modified Capabilities

<!-- none — service/chart contracts unchanged; only secret provisioning changes -->

## Impact

- **Remove:** `infra/k8s/secrets.yaml`
- **Add:** `infra/eso/` manifests, ESO operator install in Tiltfile, devenv Doppler + dotenv
- **Update:** `Tiltfile`, `devenv.nix`, `CONTRIBUTING.md`
- **Issue #10:** implementation tracking; blocks #4 ArgoCD for non-local deploy
- **Unchanged:** `refurbished-marketplace` and `kafka` Helm templates, application env/secretKeyRef usage
