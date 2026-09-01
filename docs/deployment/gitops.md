# GitOps deployment (Argo CD)

Talos is the runtime for **dev** and **prod**. Argo CD itself stays cluster bootstrap (talos-proxmox). This repo only adds a thin root Application per cluster.

## Where env lives

| Layer       | What                                       | Dev                                                                  | Prod                                                                                          |
| ----------- | ------------------------------------------ | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Cluster     | kubeconfig, apply one root                 | `~/.kube/talos-dev.yaml` + `dev-root`                                | prod kubeconfig + `prod-root`                                                                 |
| Secrets     | `operators/doppler-token` (not Helm)       | `doppler-token.dev.secret.yaml`                                      | `doppler-token.prd.secret.yaml`                                                               |
| Git (root)  | `namePrefix`, `targetRevision`, `imageTag` | `dev`, branch or `main`; images = `$ARGOCD_APP_REVISION`             | `prod`, pin git + `imageTag` SHA                                                              |
| Git (chart) | `values.yaml`                              | default (shop-dev, 1 Kafka replica, 1-day topic retention, 1 tunnel) | `values-prod.yaml` on marketplace (hosts), kafka (RF 3, 7-day retention), tunnel (2 replicas) |

Do not apply both roots on one cluster (`ecommerce` is shared). Do not put Doppler config names in Helm. Cloudflare Public Hostnames stay in Zero Trust; origin DNS is `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

Child Applications inherit `targetRevision` via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. On talos-dev, set `dev-root` `targetRevision` to the branch; `global.imageTag` is `$ARGOCD_APP_REVISION` (no per-push SHA edit). Wait for GHCR `:<sha>` or pods ImagePullBackOff until the image job finishes. Retarget to `main` before PR image cleanup.

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

Cilium is cluster-owned in **talos-proxmox**, not an Argo app. See [cilium.md](cilium.md).

`monitoring` is privileged PSS for node-exporter.

**Bootstrap:** Doppler service token Secret in `operators` — see [secrets](../development/secrets.md).

Sync waves: operators (0) → observability (1) → marketplace (3) → kafka (4) → cloudflare-tunnel (5).

```
infra/argocd/
├── app-of-apps/
│   ├── values.yaml
│   └── templates/applications.tpl
├── dev/root.yaml
└── prod/root.yaml
```

## Images

Marketplace and Kafka Connect images: `ghcr.io/phuchoang2603/refurbished-marketplace/<name>:<sha>` (and `:main` on the default branch). PRs also get `:pr-<n>`; Argo must pin the SHA. See [ci.md](ci.md).

## Related

- [cilium.md](cilium.md) — CNI, Gateway API, Cloudflare origin, Hubble
- [ci.md](ci.md) — GHCR publish and PR cleanup
