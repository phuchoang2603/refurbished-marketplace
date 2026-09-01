# GitOps deployment (Argo CD)

One Talos cluster. `talos-root` (`infra/argocd/talos/root.yaml`) → [`infra/argocd/app-of-apps`](../../infra/argocd/app-of-apps/). Chart defaults are production-like (GHCR images, full observability). Operator/Kafka/tunnel overlays still use chart-adjacent `values-staging.yaml` where those charts keep env files.

Child Applications inherit `targetRevision` via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Point the root at a branch and set `global.imageTag` to that commit SHA (after CI publishes GHCR images) to run a PR on the cluster. One live `ecommerce` namespace — one git revision at a time. Retarget to `main` before PR image cleanup.

## What Argo CD syncs

| Component                 | Source        | Pin                                                  | Namespace           |
| ------------------------- | ------------- | ---------------------------------------------------- | ------------------- |
| External Secrets Operator | Wrapper chart | upstream chart + Doppler `ClusterSecretStore`        | `operators`         |
| CloudNativePG             | Wrapper chart | upstream chart                                       | `operators`         |
| Strimzi                   | Wrapper chart | `watchAnyNamespace=true`                             | `operators`         |
| `observability`           | Wrapper chart | `victoria-metrics-k8s-stack` `0.86.0`                | `monitoring`        |
| `refurbished-marketplace` | This repo     | CNPG, ExternalSecrets, migrations, services, Gateway | `ecommerce`         |
| `kafka`                   | This repo     | Debezium reads secrets/DBs in `ecommerce`            | `kafka`             |
| `cloudflare-tunnel`       | This repo     | `cloudflared`; token via Doppler ExternalSecret      | `cloudflare-tunnel` |

Cilium is cluster-owned in **talos-proxmox** (`apps/values/cilium.yaml` + L2/Gateway manifests), not this repo and not an Argo app. See [`infra/cilium/README.md`](../../infra/cilium/README.md). Argo here may later apply marketplace policies/routes only.

`monitoring` is privileged PSS for node-exporter. Apply `talos-root` after Argo CD is installed on the cluster (Helm, not Tilt).

**Bootstrap:** Doppler service token Secret in `operators` — see [secrets](../development/secrets.md).

Sync waves: operators (0) → observability (1) → marketplace (3) → kafka (4) → cloudflare-tunnel (5).

```
infra/argocd/
├── app-of-apps/
│   ├── values.yaml
│   └── templates/applications.tpl
└── talos/root.yaml
```

## Images

Marketplace and Kafka Connect images: `ghcr.io/phuchoang2603/refurbished-marketplace/<name>:<sha>` (and `:main` on the default branch). PRs also get `:pr-<n>`; Argo must pin the SHA. See [ci.md](ci.md).

## Related

- [cilium.md](cilium.md) — CNI, Gateway API, Cloudflare origin, Hubble
- [ci.md](ci.md) — GHCR publish and PR cleanup
