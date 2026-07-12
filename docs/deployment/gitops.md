# GitOps deployment (Argo CD)

Staging syncs from `infra/argocd/staging/`. Tilt uses chart defaults locally — no dev Argo overlay.

## What Argo CD syncs

| Component                           | Source                  | Pin                                                                         | Namespace      |
| ----------------------------------- | ----------------------- | --------------------------------------------------------------------------- | -------------- |
| External Secrets Operator           | This repo wrapper chart | upstream chart `2.6.0` + Doppler `ClusterSecretStore`                       | `operators`    |
| CloudNativePG                       | This repo wrapper chart | upstream chart `0.28.3`                                                     | `operators`    |
| Strimzi                             | This repo wrapper chart | upstream chart `1.0.0`, `watchAnyNamespace=true`                            | `operators`    |
| Istio base / istiod / cni / ztunnel | This repo wrappers      | official Istio charts `1.30.2` (ambient)                                    | `istio-system` |
| `observability`                     | This repo wrapper chart | `victoria-metrics-k8s-stack` `0.86.0`                                       | `monitoring`   |
| `refurbished-marketplace`           | This repo               | CNPG, ExternalSecrets, migrations, services; `global.imageTag: main` + GHCR | `ecommerce`    |
| `kafka`                             | This repo               | same image tag/registry; Debezium reads secrets/DBs in `ecommerce`          | `kafka`        |

**Terraform (not in Git):** Argo CD.

**Bootstrap (not in Git):** Doppler service token secret in `operators` — see [secrets](../development/secrets.md).

**Marketplace chart** (`infra/charts/refurbished-marketplace`): CNPG clusters, ExternalSecrets, goose migration Jobs, and service Deployments. Staging image pins live on `infra/argocd/staging/apps/refurbished-marketplace.yaml`.

Sync order is set on Application manifests under `infra/argocd/staging/apps/`: operators + Istio base (0) → observability + istiod/cni (1) → ztunnel (2) → marketplace (3) → kafka (4).

Inside `refurbished-marketplace`, resource sync waves order work as: ExternalSecrets (2) → CNPG clusters (3) → migration Jobs (4) → Deployments / waypoint (5).

```
infra/argocd/staging/apps/
├── operators-external-secrets.yaml
├── operators-cnpg.yaml
├── operators-strimzi.yaml
├── istio-base.yaml
├── istio-istiod.yaml
├── istio-cni.yaml
├── istio-ztunnel.yaml
├── observability.yaml
├── refurbished-marketplace.yaml
└── kafka.yaml
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
