# GitOps deployment (Argo CD)

How the staging cluster syncs application manifests from this repository. Local development uses Tilt with chart defaults — there is no `dev` Argo overlay. A production environment will be added later.

## Scope

**In this repo:**

- App-of-apps manifests under `infra/argocd/staging/`
- Environment-specific Helm values embedded in each Application under `spec.source.helm.values`
- Chart image registry wiring (`global.imageRegistry` + `global.imageTag`)

**Out of scope (bootstrap, done outside Git):**

- Argo CD installation and cluster registration
- Doppler service token and `ClusterSecretStore`
- Terraform wiring that creates the root Application on the cluster

## Layout

```
infra/argocd/staging/
├── root.yaml              # Bootstrap applies this once on the staging cluster
└── apps/                  # Child Applications (Helm values inline per app)
```

The root Application syncs the `apps/` directory **without** a `destination.namespace`, so child Application CRs are created in the same namespace as the root (typically `default`). Workload Applications still set `destination.namespace` for Helm releases (`operators`, `ecommerce`).

| Sync wave | Applications                                      |
| --------- | ------------------------------------------------- |
| 0         | External Secrets Operator, CloudNativePG, Strimzi |
| 1         | `refurbished-marketplace` Helm chart              |
| 2         | `kafka` Helm chart                                |

Operator Helm chart versions match the Tiltfile (ESO 2.6.0, CNPG 0.28.3, Strimzi 1.0.0).

## Staging

| Setting   | Value                             |
| --------- | --------------------------------- |
| Cluster   | Proxmox                           |
| Image tag | `:main`                           |
| Update    | Automatic after CI push to `main` |

### Image references

Charts resolve service images via `global.imageRegistry` and `global.imageTag`:

- **Tilt:** `imageRegistry` is empty — templates use local image names unchanged.
- **Staging:** `imageRegistry: ghcr.io/phuchoang2603/refurbished-marketplace` and `imageTag: main`.

Release CI builds **all twelve** images on every run to `main`, tagging each with `:main` and `:${{ github.sha }}`.

### Payment gateway simulator

The simulator runs from the marketplace chart in staging. The marketplace Application sets:

```yaml
services:
  web:
    env:
      HOSTED_PAYMENT_BASE_URL: http://payment-gateway-simulator:8097
```

Tilt keeps `http://localhost:8097` in the base chart `values.yaml` for port-forward access.

## Debugging before merge to main

Argo CD reads chart and manifest paths from Git. While validating a change on the cluster, set `targetRevision` on repo-sourced Applications to your feature branch (for example `fix/argocd-staging-debug`) instead of `main`. After merge, set `targetRevision` back to `main` on:

- `infra/argocd/staging/root.yaml`
- `infra/argocd/staging/apps/refurbished-marketplace.yaml`
- `infra/argocd/staging/apps/kafka.yaml`

Apply the root Application to the **`default`** namespace — not `argocd` — so child Application CRs are not tied to the Argo CD install namespace.

## Bootstrap checklist

Before the first successful sync:

1. Argo CD installed on the cluster
2. Root Application applied from `infra/argocd/staging/root.yaml` into `default`
3. Doppler token secret and `ClusterSecretStore` present (see [secrets](../development/secrets.md))
4. GHCR pull access configured if images are private (`imagePullSecrets` — not automated in this repo)

## Related docs

- [secrets.md](../development/secrets.md) — Doppler and External Secrets Operator (shared with local Tilt)
- [ci.md](ci.md) — GitHub Actions and GHCR releases
