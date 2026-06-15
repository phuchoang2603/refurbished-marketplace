# GitOps deployment (Argo CD)

Staging syncs from `infra/argocd/staging/`. Tilt uses chart defaults locally — no dev Argo overlay.

## What Argo CD syncs

| Component                 | Source                  | Pin                                     | Namespace   |
| ------------------------- | ----------------------- | --------------------------------------- | ----------- |
| CloudNativePG             | Argo CD (upstream Helm) | chart `0.28.3`                          | `operators` |
| Strimzi                   | Argo CD (upstream Helm) | chart `1.0.0`, `watchAnyNamespace=true` | `operators` |
| `refurbished-marketplace` | This repo               | `global.imageTag: main` + GHCR registry | `ecommerce` |
| `kafka`                   | This repo               | same image tag/registry                 | `ecommerce` |

**Terraform (not in Git):** Argo CD, ESO (`2.6.0`), Doppler token, `ClusterSecretStore`.

Sync order: operators (wave 0) → marketplace (1) → kafka (2). Values for repo charts live inline on each Application under `spec.source.helm.values`.

```
infra/argocd/staging/
├── root.yaml
└── apps/   # operators-cnpg, operators-strimzi, refurbished-marketplace, kafka
```

## Bootstrap

1. Terraform: Argo CD, ESO, `operators` + `ecommerce` namespaces, Doppler bootstrap ([secrets](../development/secrets.md))
2. Apply `infra/argocd/staging/root.yaml` to the Argo CD namespace
3. GHCR pull access if images are private

## Branch debugging

Point `targetRevision` at a feature branch on repo-sourced Applications (`root.yaml`, `refurbished-marketplace.yaml`, `kafka.yaml`) to test before merge; set back to `main` after. Operator apps pin upstream chart versions only.

## See also

- [ci.md](ci.md) — image builds and GHCR tags
- [secrets.md](../development/secrets.md) — Doppler + ESO
