# GitOps deployment (Argo CD)

Staging syncs from `infra/argocd/staging/`. Local Colima syncs from `infra/argocd/local/` (same GitOps model; omits observability).

## What Argo CD syncs

| Component                           | Source                  | Pin                                                              | Namespace           |
| ----------------------------------- | ----------------------- | ---------------------------------------------------------------- | ------------------- |
| External Secrets Operator           | This repo wrapper chart | upstream chart `2.6.0` + Doppler `ClusterSecretStore`            | `operators`         |
| CloudNativePG                       | This repo wrapper chart | upstream chart `0.28.3`                                          | `operators`         |
| Strimzi                             | This repo wrapper chart | upstream chart `1.0.0`, `watchAnyNamespace=true`                 | `operators`         |
| Istio base / istiod / cni / ztunnel | This repo wrappers      | official Istio charts `1.30.2` (ambient)                         | `istio-system`      |
| `observability`                     | This repo wrapper chart | `victoria-metrics-k8s-stack` `0.86.0` (staging only)             | `monitoring`        |
| `refurbished-marketplace`           | This repo               | CNPG, ExternalSecrets, migrations, services                      | `ecommerce`         |
| `kafka`                             | This repo               | Debezium reads secrets/DBs in `ecommerce`                        | `kafka`             |
| `cloudflare-tunnel`                 | This repo               | `cloudflared` connector; tunnel token via Doppler ExternalSecret | `cloudflare-tunnel` |

**Local (Colima):** `infra/argocd/local/` uses chart `values.yaml` (k3s CNI, ambient + `.dev` ingress, short image names) plus Cloudflare Tunnel. `bootstrap-local-argocd` pins Applications to the current git branch. See [local-setup](../development/local-setup.md).

**Staging:** `values-staging.yaml` overlays (GHCR `:main`, production hostnames, RKE2 CNI paths, observability).

**Terraform (not in Git for staging):** Argo CD on remote clusters.

**Bootstrap (not in Git):** Doppler service token secret in `operators` — see [secrets](../development/secrets.md).

**Marketplace chart** (`infra/charts/refurbished-marketplace`): CNPG clusters, ExternalSecrets, goose migration Jobs, and service Deployments. Staging overlays live in `values-staging.yaml` (wired via `valueFiles`).

Sync waves under `infra/argocd/staging/apps/` (and the same waves under `infra/argocd/local/apps/`): operators + Istio base (0) → observability + istiod/cni (1) → ztunnel (2) → marketplace (3) → kafka (4) → cloudflare-tunnel (5). Local omits observability only.

Inside `refurbished-marketplace`, resource sync waves order work as: ExternalSecrets (2) → CNPG clusters (3) → migration Jobs (4) → Deployments / waypoint / ingress Gateway (5) → HTTPRoutes (6).

```
infra/argocd/
├── local/
│   ├── root.yaml
│   └── apps/          # operators, Istio, marketplace, kafka, cloudflare-tunnel
└── staging/
    ├── root.yaml
    └── apps/          # + observability; staging valueFiles / image tags
```

## Bootstrap

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
