# GitOps deployment (Argo CD)

Talos **dev** and **prod** are the runtimes. Argo CD runs on the **gpu** cluster (talos-proxmox `bootstrap-gpu.sh`) and registers remotes named `dev` and `prod`. This repo’s root Applications live in gpu `argo-cd`; children destine those cluster names.

## Where env lives

| Layer       | What                                       | Dev                                                                  | Prod                                                                                          |
| ----------- | ------------------------------------------ | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Argo        | apply roots                                | `~/.kube/talos-gpu.yaml` + `dev-root` (`destinationName: dev`)       | same gpu kubeconfig + `prod-root` (`destinationName: prod`)                                   |
| Workloads   | kubeconfig for Doppler / kubectl           | `~/.kube/talos-dev.yaml`                                             | prod kubeconfig                                                                               |
| Secrets     | `operators/doppler-token` (not Helm)       | `doppler-token.dev.secret.yaml` on talos-dev                         | `doppler-token.prd.secret.yaml` on prod                                                       |
| Git (root)  | `namePrefix`, `targetRevision`, `imageTag` | `dev`, branch or `main`; images = `$ARGOCD_APP_REVISION`             | `prod`, git `main`; images = `:main`                                                          |
| Git (chart) | `values.yaml`                              | default (shop-dev, 1 Kafka replica, 1-day topic retention, 1 tunnel) | `values-prod.yaml` on marketplace (hosts), kafka (RF 3, 7-day retention), tunnel (2 replicas) |

Both roots may be applied on gpu; they target different clusters. Do not put Doppler config names in Helm. Cloudflare Public Hostnames stay in Zero Trust; origin DNS is `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

Child Applications inherit `targetRevision` via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Change `dev-root` `spec.source.targetRevision` in `infra/argocd/dev/root.yaml`, commit, and `kubectl apply -f` that file on gpu. `global.imageTag` is `$ARGOCD_APP_REVISION`. Wait for GHCR `:<sha>` or pods ImagePullBackOff until the image job finishes. Leave `targetRevision` on the branch you are running until you deliberately retarget (do not flip it to `main` just because a PR merged).

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

Marketplace and Kafka Connect images: `ghcr.io/phuchoang2603/refurbished-marketplace/<name>:<sha>` on talos-dev, or `:main` on prod. See [ci.md](ci.md).

## Related

- [cilium.md](cilium.md) — CNI, Gateway API, Cloudflare origin
- [ci.md](ci.md) — GHCR `:main` / `:<sha>` and PR SHA cleanup
