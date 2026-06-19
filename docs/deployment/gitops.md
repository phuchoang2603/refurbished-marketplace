# GitOps deployment (Argo CD)

Staging syncs from `infra/argocd/staging/`. Tilt uses chart defaults locally — no dev Argo overlay.

## What Argo CD syncs

| Component                       | Source                  | Pin                                      | Namespace   |
| ------------------------------- | ----------------------- | ---------------------------------------- | ----------- |
| CloudNativePG                   | Argo CD (upstream Helm) | chart `0.28.3`                           | `operators` |
| Strimzi                         | Argo CD (upstream Helm) | chart `1.0.0`, `watchAnyNamespace=true`  | `operators` |
| `refurbished-marketplace-infra` | This repo               | CNPG, ExternalSecrets, schema migrations | `ecommerce` |
| `refurbished-marketplace`       | This repo               | `global.imageTag: main` + GHCR           | `ecommerce` |
| `kafka`                         | This repo               | same image tag/registry                  | `ecommerce` |

**Terraform (not in Git):** Argo CD, ESO (`2.6.0`), Doppler token, `ClusterSecretStore`.

**Infra chart** (`infra/charts/refurbished-marketplace-infra`): CNPG clusters, ExternalSecrets, and goose migration Jobs. Staging image pins live on `infra/argocd/staging/apps/refurbished-marketplace-infra.yaml`.

Sync order is set on Application manifests under `infra/argocd/staging/apps/`: operators (0) → infra (1) → marketplace (2) → kafka (3).

```
infra/argocd/staging/apps/
├── operators-cnpg.yaml
├── operators-strimzi.yaml
├── refurbished-marketplace-infra.yaml
├── refurbished-marketplace.yaml
└── kafka.yaml
```

## Bootstrap

1. Terraform: Argo CD, ESO, `operators` + `ecommerce` namespaces, Doppler bootstrap ([secrets](../development/secrets.md))
2. Apply `infra/argocd/staging/root.yaml` to the Argo CD namespace
3. GHCR pull access if images are private

## See also

- [ci.md](ci.md) — image builds and GHCR tags
- [secrets.md](../development/secrets.md) — Doppler + ESO
