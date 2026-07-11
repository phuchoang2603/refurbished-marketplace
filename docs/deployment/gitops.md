# GitOps deployment (Argo CD)

Staging syncs from `infra/argocd/staging/`. Tilt uses chart defaults locally — no dev Argo overlay.

## What Argo CD syncs

| Component                       | Source                  | Pin                                                   | Namespace    |
| ------------------------------- | ----------------------- | ----------------------------------------------------- | ------------ |
| External Secrets Operator       | This repo wrapper chart | upstream chart `2.6.0` + Doppler `ClusterSecretStore` | `operators`  |
| CloudNativePG                   | This repo wrapper chart | upstream chart `0.28.3`                               | `operators`  |
| Strimzi                         | This repo wrapper chart | upstream chart `1.0.0`, `watchAnyNamespace=true`      | `operators`  |
| `observability`                 | This repo wrapper chart | `victoria-metrics-k8s-stack` `0.86.0`                 | `monitoring` |
| `refurbished-marketplace-infra` | This repo               | CNPG, ExternalSecrets, schema migrations              | `ecommerce`  |
| `refurbished-marketplace`       | This repo               | `global.imageTag: main` + GHCR                        | `ecommerce`  |
| `kafka`                         | This repo               | same image tag/registry                               | `ecommerce`  |

**Terraform (not in Git):** Argo CD.

**Bootstrap (not in Git):** Doppler service token secret in `operators` — see [secrets](../development/secrets.md).

**Infra chart** (`infra/charts/refurbished-marketplace-infra`): CNPG clusters, ExternalSecrets, and goose migration Jobs. Staging image pins live on `infra/argocd/staging/apps/refurbished-marketplace-infra.yaml`.

Sync order is set on Application manifests under `infra/argocd/staging/apps/`: operators (0) → observability + infra (1) → marketplace (2) → kafka (3).

```
infra/argocd/staging/apps/
├── operators-external-secrets.yaml
├── operators-cnpg.yaml
├── operators-strimzi.yaml
├── observability.yaml
├── refurbished-marketplace-infra.yaml
├── refurbished-marketplace.yaml
└── kafka.yaml
```

## Bootstrap

1. Terraform: Argo CD, `operators` + `ecommerce` namespaces
2. Add `prd` application secrets in Doppler ([secrets](../development/secrets.md))
3. Apply Doppler bootstrap secret: `kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml`
4. Apply `infra/argocd/staging/root.yaml` to the Argo CD namespace
5. GHCR pull access if images are private

## See also

- [ci.md](ci.md) — image builds and GHCR tags
- [secrets.md](../development/secrets.md) — Doppler + ESO
- [observability.md](../observability.md) — VictoriaMetrics stack and Grafana access
