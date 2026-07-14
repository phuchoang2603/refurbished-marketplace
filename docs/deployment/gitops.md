# GitOps deployment (Argo CD)

Staging syncs everything from `infra/argocd/staging/`. Local Colima: Tilt owns the marketplace chart; Argo (`infra/argocd/local/`) owns operators, Istio, Kafka, observability, and Cloudflare Tunnel.

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

**Local (Colima):** `tilt up` installs Argo CD, applies `infra/argocd/local/` (pinned to the current git branch), and deploys the marketplace chart via Tilt. Chart `values.yaml` is Colima-local (k3s CNI, ambient + `.dev` ingress, short image names). See [local-setup](../development/local-setup.md).

**Staging:** Full app-of-apps including marketplace, with `values-staging.yaml` overlays (GHCR `:main`, production hostnames, RKE2 CNI paths).

**Terraform (not in Git for staging):** Argo CD on remote clusters.

**Bootstrap (not in Git):** Doppler service token secret in `operators` — see [secrets](../development/secrets.md). Locally Tilt applies the `dev` Doppler secret.

**Marketplace chart** (`infra/charts/refurbished-marketplace`): CNPG clusters, ExternalSecrets, goose migration Jobs, and service Deployments. Staging overlays live in `values-staging.yaml` (wired via `valueFiles`).

Staging sync waves: operators + Istio base (0) → observability + istiod/cni (1) → ztunnel (2) → marketplace (3) → kafka (4) → cloudflare-tunnel (5). Local Argo uses the same waves for infra apps (no marketplace Application).

Inside `refurbished-marketplace`, resource sync waves (staging Argo) order work as: ExternalSecrets (2) → CNPG clusters (3) → migration Jobs (4) → Deployments / waypoint / ingress Gateway (5) → HTTPRoutes (6).

```
infra/argocd/
├── local/
│   ├── root.yaml
│   └── apps/          # operators, Istio, observability, kafka, cloudflare-tunnel (no marketplace)
└── staging/
    ├── root.yaml
    └── apps/          # + marketplace; staging valueFiles / image tags
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
