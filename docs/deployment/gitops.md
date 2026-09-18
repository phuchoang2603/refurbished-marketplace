# GitOps deployment (Argo CD)

Talos **dev** and **prod** run the workloads. Argo CD runs on the **management** cluster and registers those workload clusters as `dev` and `prod`. Two independent roots are applied to Argo CD:

1. `platform-dev` / `platform-prod` from [`talos-proxmox`](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps) install shared operators and observability.
2. `dev-root` / `prod-root` from this repository deploy marketplace-owned resources that consume those platform APIs.

## Fresh installation

Provision the Talos clusters and Argo CD with `talos-proxmox`, then register the workload clusters. Before applying a marketplace root, verify the matching platform root is healthy and the destination cluster has:

- a default StorageClass;
- External Secrets, CloudNativePG, Strimzi, MongoDB Community, and VictoriaMetrics operators with their CRDs;
- `operators/doppler-token` and a Ready `ClusterSecretStore/doppler`;
- Grafana, VMAgent, VLAgent, VictoriaLogs, and VictoriaTraces in `monitoring`;
- the marketplace container images required by the selected Git revision in GHCR.

Apply the platform root from the sibling checkout first:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-argocd.yaml" apply --server-side \
  -f ../talos-proxmox/apps/argocd/roots/dev.yaml
```

Prepare the application secrets with the workload kubeconfig as described in [secrets.md](../development/secrets.md), then apply this repository's root on the management cluster:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-argocd.yaml" apply --server-side \
  -f infra/argocd/dev/root.yaml
```

Use `apps/argocd/roots/prod.yaml` and `infra/argocd/prod/root.yaml` for production. Both repositories commit their roots to `main`.

## Environment configuration

| Layer            | Dev                                        | Prod                                        |
| ---------------- | ------------------------------------------ | ------------------------------------------- |
| Destination      | registered Argo cluster `dev`              | registered Argo cluster `prod`              |
| Platform root    | `platform-dev`                             | `platform-prod`                             |
| Marketplace root | `dev-root`                                 | `prod-root`                                 |
| Doppler token    | `operators/doppler-token` for config `dev` | `operators/doppler-token` for config `prd`  |
| Images           | `:<git-sha>` via `$ARGOCD_APP_REVISION`    | rolling `:main`                             |
| Values           | chart `values.yaml`                        | chart `values.yaml` plus `values-prod.yaml` |

Child Applications inherit the root Git revision through `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Dev converts that revision to the immutable image tag `$ARGOCD_APP_REVISION`; wait for all required `:<sha>` images before syncing. Production uses `:main`. Doppler config names do not belong in Helm or Argo values.

Cloudflare Public Hostnames remain in Zero Trust. The shop/pay origin is `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

## Marketplace ownership

| Application               | Marketplace-owned resources                                                                                 | Namespace                               |
| ------------------------- | ----------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| `mongodb`                 | `MongoDBCommunity`, credentials, workload RBAC, Cilium policy                                               | `ecommerce`                             |
| `meilisearch`             | Meilisearch workload/PVC, credentials, Cilium policy                                                        | `ecommerce`                             |
| `refurbished-marketplace` | CNPG Clusters, ExternalSecrets, migrations, services, Gateway/HTTPRoutes, VMPodScrape, dashboard ConfigMaps | `ecommerce`, dashboards in `monitoring` |
| `kafka`                   | Kafka/NodePool, topics, Connect/connectors, secret-reader RBAC, UI                                          | `kafka`, RBAC in `ecommerce`            |
| `cloudflare-tunnel`       | cloudflared and its ExternalSecret                                                                          | `cloudflare-tunnel`                     |

The matching platform root owns the operators, CRDs, `ClusterSecretStore/doppler`, the `monitoring` namespace, and the complete Victoria stack. Cilium, Gateway API installation, and storage are also cluster-owned in `talos-proxmox`.

MongoDB is the products catalog source of truth; Meilisearch is its storefront projection. PostgreSQL schema migrations still initialize new empty databases before their services start.

## Ordering and health

Marketplace child annotations give this local order:

```text
MongoDB + Meilisearch (2) → marketplace (3) → Kafka (4) → cloudflared (5)
```

These waves order resources within the marketplace root. They do not order the independent platform root, and the current Argo CD configuration does not restore `Application` CR health assessment for child-readiness orchestration. Verify platform readiness explicitly before applying the marketplace root.

Useful checks:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-argocd.yaml" get applications -n argo-cd
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get storageclass
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get clustersecretstore doppler
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get crd | grep -E 'cnpg|strimzi|mongodb|external-secrets|victoriametrics'
```

## Repository layout

```text
infra/argocd/
├── app-of-apps/
│   ├── values.yaml
│   └── templates/applications.tpl
├── dev/root.yaml
└── prod/root.yaml
```

See [ci.md](ci.md) for image publication, [cilium.md](cilium.md) for networking, [observability.md](observability.md) for the telemetry contract, and the [`talos-proxmox` application guide](https://github.com/phuchoang2603/talos-proxmox/blob/main/apps/README.md) for platform installation and administration.
