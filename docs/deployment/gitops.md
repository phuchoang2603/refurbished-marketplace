# GitOps deployment (Argo CD)

Talos **dev** and **prod** each run their own Argo CD in namespace `argo-cd`, installed by the [`talos-proxmox`](https://github.com/phuchoang2603/talos-proxmox) platform OpenTofu root. Each Argo CD manages only its local cluster and hosts two independent roots:

1. `platform` from `talos-proxmox` installs shared operators, the Doppler `ClusterSecretStore`, storage, observability, and Cloudflare Tunnel.
2. `dev-root` / `prod-root` from this repository deploy marketplace-owned resources that consume those platform APIs.

All marketplace Applications destine `https://kubernetes.default.svc`. There are no cluster registrations or cross-environment kubeconfigs.

## Fresh installation

Bring up the environment with `talos-proxmox` first; a successful platform apply means Argo CD and the `platform` root are installed. The marketplace root can be applied immediately afterwards: its children retry with backoff until platform operators and CRDs exist.

Fetch the environment kubeconfig from Doppler (the project must be explicit because `devenv.nix` defaults to `refurbished-marketplace`):

```bash
export CLUSTER_ENV=dev  # use prod for production
umask 077
doppler secrets get KUBECONFIG --plain \
  --project talos-proxmox --config "$CLUSTER_ENV" \
  > "$HOME/.kube/talos-${CLUSTER_ENV}.yaml"
export KUBECONFIG="$HOME/.kube/talos-${CLUSTER_ENV}.yaml"
```

Apply the Doppler token as described in [secrets.md](../development/secrets.md), then the project and root with the same kubeconfig:

```bash
kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.dev.secret.yaml
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply --server-side -f infra/argocd/dev/root.yaml
```

For production use `doppler-token.prd.secret.yaml` and `infra/argocd/prod/root.yaml`. Both roots commit `main`. Dev needs every required GHCR `:<sha>` image for the selected revision before its children can become Healthy.

## Environment configuration

| Layer            | Dev                                        | Prod                                        |
| ---------------- | ------------------------------------------ | ------------------------------------------- |
| Argo CD          | dev's own, UI `http://10.69.11.254`        | prod's own, UI `http://10.69.12.254`        |
| Kubeconfig       | Doppler `talos-proxmox/dev` `KUBECONFIG`   | Doppler `talos-proxmox/prod` `KUBECONFIG`   |
| Marketplace root | `dev-root`                                 | `prod-root`                                 |
| Doppler token    | `ecommerce/doppler-token` for config `dev` | `ecommerce/doppler-token` for config `prd`  |
| Images           | `:<git-sha>` via `$ARGOCD_APP_REVISION`    | rolling `:main`                             |
| Values           | chart `values.yaml`                        | chart `values.yaml` plus `values-prod.yaml` |

Child Applications inherit the root Git revision through `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Dev converts that revision to the immutable image tag `$ARGOCD_APP_REVISION`; production uses `:main`. Doppler config names do not belong in Helm or Argo values.

Cloudflare Public Hostnames remain in Zero Trust. The shop/pay origin is `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

## Marketplace ownership

| Application               | Marketplace-owned resources                                                        | Namespace                    |
| ------------------------- | ---------------------------------------------------------------------------------- | ---------------------------- |
| `secret-store`            | `SecretStore/doppler`, the single owner shared by every marketplace ExternalSecret | `ecommerce`                  |
| `mongodb`                 | `MongoDBCommunity`, credentials, workload RBAC, Cilium policy                      | `ecommerce`                  |
| `meilisearch`             | Meilisearch workload/PVC, credentials, Cilium policy                               | `ecommerce`                  |
| `refurbished-marketplace` | CNPG Clusters, ExternalSecrets, migrations, services, Gateway/HTTPRoutes           | `ecommerce`                  |
| `kafka`                   | Kafka/NodePool, topics, Connect/connectors, secret-reader RBAC, UI                 | `kafka`, RBAC in `ecommerce` |
| `cloudflare-tunnel`       | platform-owned in `talos-proxmox`                                                  | `cloudflare-tunnel`          |

The `platform` root owns the operators, CRDs, cloudflared, and the telemetry pipeline (`otel-agent` in both environments; ClickHouse and HyperDX on prod). Marketplace workloads only export OTLP to `otel-agent`; see [observability.md](observability.md). The marketplace `secret-store` Application owns `SecretStore/doppler` in `ecommerce`; no other chart renders it. Cilium, Gateway API CRDs, and storage are also owned by `talos-proxmox`.

MongoDB is the products catalog source of truth; Meilisearch is its storefront projection. PostgreSQL schema migrations still initialize new empty databases before their services start.

PostgreSQL services and migration jobs use the same `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, and `PGSSLMODE` settings. Helm reads each database password from its service secret; neither service code nor the goose connection string embeds or interpolates the password. Set these PostgreSQL environment variables when connecting locally as well.

## Ordering and convergence

Marketplace child annotations give this local order:

```text
secret-store (1) → MongoDB + Meilisearch (2) → marketplace (3) → Kafka (4)
```

Waves order submission of child Applications only; they do not wait for each child to become Healthy and do not order the marketplace root against `platform`. Children converge asynchronously: unlimited retries with backoff (capped at five minutes) and `SkipDryRunOnMissingResource` absorb CRDs, storage, and secret stores that platform Applications have not finished installing. The only manual prerequisite is `ecommerce/doppler-token`.

Useful checks:

```bash
kubectl get applications -n argo-cd
kubectl get storageclass
kubectl get secretstore doppler -n ecommerce
kubectl get externalsecrets -n ecommerce
kubectl get crd | grep -E 'cnpg|strimzi|mongodb|external-secrets'
kubectl get svc otel-agent -n observability
```

A child stuck retrying shows its last sync error in the environment's Argo CD UI. A missing or invalid token shows up on `<env>-secret-store` first.

## Repository layout

```text
infra/argocd/
├── app-of-apps/
│   ├── values.yaml
│   └── templates/applications.tpl
├── project.yaml
├── dev/root.yaml
└── prod/root.yaml
```

See [ci.md](ci.md) for image publication, [cilium.md](cilium.md) for networking, and [observability.md](observability.md) for the telemetry contract. Platform details live in `talos-proxmox`'s [GitOps architecture](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/architecture/gitops.md) and [cluster access](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/operations/cluster-access.md) guides.
