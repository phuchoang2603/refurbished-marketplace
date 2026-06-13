# GitOps deployment (Argo CD)

How staging and production clusters sync application manifests from this repository. Local development uses Tilt with chart defaults — there is no `dev` Argo overlay.

## Scope

**In this repo:**

- App-of-apps manifests under `infra/argocd/`
- Environment-specific Helm values embedded in each Application under `spec.source.helm.values`
- Chart image registry wiring (`global.imageRegistry` + `global.imageTag`)

**Out of scope (bootstrap, done outside Git):**

- Argo CD installation and cluster registration
- Doppler service token and `ClusterSecretStore`
- Terraform wiring that creates the root Application per cluster

## Layout

```
infra/argocd/
├── staging/
│   ├── root.yaml              # Terraform applies this on the staging cluster
│   └── apps/                  # Child Applications (Helm values inline per app)
├── production/
│   ├── root.yaml
│   └── apps/
```

Each environment uses a **root Application** that syncs the `apps/` directory into the `argocd` namespace. Child Applications deploy:

| Sync wave | Applications                                      |
| --------- | ------------------------------------------------- |
| 0         | External Secrets Operator, CloudNativePG, Strimzi |
| 1         | `refurbished-marketplace` Helm chart              |
| 2         | `kafka` Helm chart                                |

Operator Helm chart versions match the Tiltfile (ESO 2.6.0, CNPG 0.28.3, Strimzi 1.0.0).

## Environments

| Environment  | Cluster | Image tag                       | Update mechanism                                 |
| ------------ | ------- | ------------------------------- | ------------------------------------------------ |
| Local (Tilt) | devenv  | Short names (`web`, `users`, …) | `tilt up` rebuilds images                        |
| Staging      | Proxmox | `:main`                         | Automatic after CI push to `main`                |
| Production   | AWS     | Commit SHA                      | Manual Git edit to production Application values |

### Image references

Charts resolve service images via `global.imageRegistry` and `global.imageTag`:

- **Tilt:** `imageRegistry` is empty — templates use local image names unchanged.
- **Staging / production:** `imageRegistry: ghcr.io/phuchoang2603/refurbished-marketplace` and a shared tag for all services.

Release CI builds **all twelve** images on every run to `main`, tagging each with `:main` and `:${{ github.sha }}`. Production promotion sets one SHA in both marketplace and kafka Application manifests so every service pulls the same release.

### Promoting to production

1. Confirm the target commit SHA was built by [Release Images](../../.github/workflows/release-images.yml) (check GHCR tags).
2. Edit `global.imageTag` in both:
   - `infra/argocd/production/apps/refurbished-marketplace.yaml`
   - `infra/argocd/production/apps/kafka.yaml`
   ```yaml
   global:
     imageTag: abc123def456 # commit SHA
   ```
3. Merge to `main`. Argo CD syncs production automatically (or trigger sync manually).

To roll back, revert the `imageTag` change or use Argo CD history.

### Payment gateway simulator

The simulator runs from the marketplace chart in staging and production. The marketplace Application sets:

```yaml
services:
  web:
    env:
      HOSTED_PAYMENT_BASE_URL: http://payment-gateway-simulator:8097
```

Tilt keeps `http://localhost:8097` in the base chart `values.yaml` for port-forward access.

## Bootstrap checklist

Before the first successful sync:

1. Argo CD installed on the cluster
2. Root Application applied (`infra/argocd/staging/root.yaml` or `production/root.yaml`)
3. Doppler token secret and `ClusterSecretStore` present (see [secrets.md](secrets.md))
4. GHCR pull access configured if images are private (`imagePullSecrets` — not automated in this repo)

## Verifying manifests locally

Environment values live in `spec.source.helm.values` on each Application. For a quick render matching staging:

```bash
# Default (Tilt) — local image names
helm template refurbished-marketplace infra/charts/refurbished-marketplace

# Staging-equivalent (values from staging Applications)
helm template refurbished-marketplace infra/charts/refurbished-marketplace \
  --set global.imageRegistry=ghcr.io/phuchoang2603/refurbished-marketplace \
  --set global.imageTag=main

helm template ecommerce-kafka-cluster infra/charts/kafka \
  --set global.imageRegistry=ghcr.io/phuchoang2603/refurbished-marketplace \
  --set global.imageTag=main
```

For a full match including nested keys (for example `HOSTED_PAYMENT_BASE_URL`), copy the `values:` block from the Application into a temp file and pass `-f`.

## Related docs

- [secrets.md](secrets.md) — Doppler and External Secrets Operator
- [ci.md](ci.md) — GitHub Actions and GHCR releases
