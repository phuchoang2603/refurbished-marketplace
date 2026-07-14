# GitOps deployment (Argo CD)

Staging syncs everything via `staging-root` → `infra/argocd/app-of-apps` + `values-staging.yaml`. Local Colima: Tilt owns the marketplace chart; Argo (`local-root` → same chart + `values-local.yaml`) owns operators, Istio, Kafka, observability, and Cloudflare Tunnel.

## What Argo CD syncs

| Component                           | Source                  | Pin                                                              | Namespace           | Local | Staging |
| ----------------------------------- | ----------------------- | ---------------------------------------------------------------- | ------------------- | ----- | ------- |
| External Secrets Operator           | This repo wrapper chart | upstream chart `2.6.0` + Doppler `ClusterSecretStore`            | `operators`         | yes   | yes     |
| CloudNativePG                       | This repo wrapper chart | upstream chart `0.28.3`                                          | `operators`         | yes   | yes     |
| Strimzi                             | This repo wrapper chart | upstream chart `1.0.0`, `watchAnyNamespace=true`                 | `operators`         | yes   | yes     |
| Istio base / istiod / cni / ztunnel | This repo wrappers      | official Istio charts `1.30.2` (ambient)                         | `istio-system`      | yes   | yes     |
| `observability`                     | This repo wrapper chart | `victoria-metrics-k8s-stack` `0.86.0`                            | `monitoring`        | yes   | yes     |
| `refurbished-marketplace`           | This repo               | CNPG, ExternalSecrets, migrations, services                      | `ecommerce`         | Tilt  | yes     |
| `kafka`                             | This repo               | Debezium reads secrets/DBs in `ecommerce`                        | `kafka`             | yes   | yes     |
| `cloudflare-tunnel`                 | This repo               | `cloudflared` connector; tunnel token via Doppler ExternalSecret | `cloudflare-tunnel` | yes   | yes     |

**Local (Colima):** `tilt up` installs Argo CD and applies `local-root` (current git branch). Both environments render children from the shared Helm chart [`infra/argocd/app-of-apps/`](../../infra/argocd/app-of-apps/); env-specific settings live inline on each root Application’s `helm.values`. Children inherit `targetRevision` via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Local omits the marketplace Application (Tilt owns that chart). See [local-setup](../development/local-setup.md).

**Staging:** Same chart; `staging-root` sets `global.imageRegistry` / `imageTag`, enables marketplace, and points chart `valueFiles` at chart-adjacent `values-staging.yaml` overlays where needed.

**Terraform (not in Git for staging):** Argo CD on remote clusters.

**Bootstrap (not in Git):** Doppler service token secret in `operators` — see [secrets](../development/secrets.md). Locally Tilt applies the `dev` Doppler secret.

**Marketplace chart** (`infra/charts/refurbished-marketplace`): CNPG clusters, ExternalSecrets, goose migration Jobs, and service Deployments. Staging overlays live in `values-staging.yaml` (wired via `valueFiles`).

Staging sync waves: operators + Istio base (0) → observability + istiod/cni (1) → ztunnel (2) → marketplace (3) → kafka (4) → cloudflare-tunnel (5). Local Argo uses the same waves for infra apps (no marketplace Application).

Inside `refurbished-marketplace`, resource sync waves (staging Argo) order work as: ExternalSecrets (2) → CNPG clusters (3) → migration Jobs (4) → Deployments / waypoint / ingress Gateway (5) → HTTPRoutes (6).

```
infra/argocd/
├── app-of-apps/              # shared child Application catalog (one template)
│   ├── values.yaml           # defaults + apps map
│   └── templates/applications.yaml
├── local/root.yaml           # inline helm.values (local)
└── staging/root.yaml         # inline helm.values (staging + global images)
```

## Bootstrap (staging)

1. Terraform: Argo CD, `operators` + `ecommerce` namespaces (`kafka` is created by the Kafka Application via `CreateNamespace=true`)
2. Add `prd` application secrets in Doppler ([secrets](../development/secrets.md))
3. Apply Doppler bootstrap secret: `kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml`
4. Apply `infra/argocd/staging/root.yaml` to the Argo CD namespace
5. GHCR pull access if images are private

## See also

- [ci.md](ci.md) — image builds and GHCR tags
- [secrets.md](../development/secrets.md) — Doppler + ESO
- [observability.md](../observability.md) — VictoriaMetrics stack and Grafana access
- [istio.md](istio.md) — ambient mesh enrollment and rollback
