# GitOps deployment (Argo CD)

The Talos **prod** cluster is the development server. It runs Argo CD in namespace `argo-cd`, installed by the [`talos-proxmox`](https://github.com/phuchoang2603/talos-proxmox) platform OpenTofu root. Argo CD manages its local cluster and hosts two independent roots:

1. `platform` from `talos-proxmox` installs shared operators, the Doppler `ClusterSecretStore`, storage, observability, and the Cloudflare operator with its `ClusterTunnel`.
2. `prod-root` from this repository deploys marketplace-owned resources that consume those platform APIs.

All marketplace Applications destine `https://kubernetes.default.svc`. There are no cluster registrations or cross-environment kubeconfigs.

## Fresh installation

Bring up the environment with `talos-proxmox` first; a successful platform apply means Argo CD and the `platform` root are installed. The marketplace root can be applied immediately afterwards: its children retry with backoff until platform operators and CRDs exist.

Fetch the environment kubeconfig from Doppler (the project must be explicit because `devenv.nix` defaults to `refurbished-marketplace`):

```bash
umask 077
doppler secrets get KUBECONFIG --plain \
  --project talos-proxmox --config prod \
  > "$HOME/.kube/talos-prod.yaml"
export KUBECONFIG="$HOME/.kube/talos-prod.yaml"
```

Apply the Doppler token as described in [secrets.md](../development/secrets.md), then the project and root with the same kubeconfig:

```bash
kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply --server-side -f infra/argocd/prod/root.yaml
```

## Environment configuration

| Layer            | Prod                                         |
| ---------------- | -------------------------------------------- |
| Argo CD          | UI `http://10.69.12.254`                     |
| Kubeconfig       | Doppler `talos-proxmox/prod` `KUBECONFIG`    |
| Marketplace root | `prod-root`                                  |
| Doppler token    | `ecommerce/doppler-token` for config `prd`   |
| Images           | Immutable SHA pinned by the release workflow |
| Values           | Chart `values.yaml` plus `values-prod.yaml`  |

The committed `infra/argocd/prod/root.yaml` tracks `main` for both the root and child Applications. The release workflow pins immutable image SHAs in `values-prod.yaml` after building images. Doppler config names do not belong in Helm or Argo values.

The marketplace chart's `TunnelBinding` publishes shop through the platform's `ClusterTunnel/talos-proxmox`; the Cloudflare operator creates its DNS records. The origin is `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`. See [networking.md](networking.md#edge).

Production Checkout uses `PAYMENT_GATEWAY_URL=http://edge.payment-gateway.svc.cluster.local:8080`, `CHECKOUT_RETURN_BASE_URL=https://shop.phuchoang.sbs`, and the two keys from `checkout-payment-gateway`; see [secrets.md](../development/secrets.md). The API key identifies the integration at Edge. Chart defaults leave the gateway disabled and refuse checkout submission. The marketplace no longer renders Payment, its database, or the hosted-page simulator, and publishes no `pay` / `pay-dev` hostname.

Production cluster end-to-end verification is deferred. Provision the gateway integration and matching credentials before deploying this release; see [the rollout gate](../order-placement.md#rollout-gate). Local Helm renders do not satisfy that gate.

## Marketplace ownership

| Application               | Marketplace-owned resources                                                                                                                              | Namespace                    |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| root (`prod-root`)        | `ecommerce` and `kafka` Namespaces with `istio.io/dataplane-mode: ambient`                                                                               | cluster                      |
| `secret-store`            | `SecretStore/doppler`, the single owner shared by every marketplace ExternalSecret                                                                       | `ecommerce`                  |
| `mongodb`                 | `MongoDBCommunity`, credentials, workload RBAC                                                                                                           | `ecommerce`                  |
| `meilisearch`             | Meilisearch workload/PVC, credentials                                                                                                                    | `ecommerce`                  |
| `refurbished-marketplace` | CNPG Clusters, ExternalSecrets, migrations, services, Gateways/HTTPRoutes, TunnelBinding, PeerAuthentication, DestinationRules, gateway webhook policies | `ecommerce`                  |
| `kafka`                   | Kafka/NodePool, topics, Connect/connectors, secret-reader RBAC, UI, PeerAuthentication                                                                   | `kafka`, RBAC in `ecommerce` |
| `cloudflare-tunnel`       | platform-owned in `talos-proxmox`                                                                                                                        | `cloudflare-operator-system` |

The `platform` root owns the operators, CRDs, the Cloudflare tunnel and its cloudflared, and the telemetry pipeline (VMAgent, VTAgent, VLAgent, Victoria stores and Grafana on prod). Marketplace workloads only export OTLP/HTTP to VMAgent and VTAgent; see [observability.md](observability.md). The marketplace `secret-store` Application owns `SecretStore/doppler` in `ecommerce`; no other chart renders it. Cilium, Istio, Gateway API CRDs, and storage are also owned by `talos-proxmox`. Child Applications do not use `CreateNamespace`; only the root creates and labels the marketplace namespaces.

MongoDB is the products catalog source of truth; Meilisearch is its storefront projection. PostgreSQL schema migrations still initialize new empty databases before their services start.

Cart and Checkout use separate persistent `Redis` resources (`cart-redis` and `checkout-redis`) from the platform's OT-CONTAINER-KIT Redis operator. Both use AOF and `noeviction`; the checkout projection also recovers its account facts and checkpoint from the retained Users topic. Deploy the Users migration and connector before enabling the new checkout image, and wait for account replay before the checkout smoke test.

Fresh installations start with empty cart and checkout Redis instances; no cart-data copy or legacy sidecar cutover is required. The release workflow pins matching image SHAs in the production overlays after publishing images.

PostgreSQL services and migration jobs use the same `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, and `PGSSLMODE` settings. Helm reads each database password from its service secret; neither service code nor the goose connection string embeds or interpolates the password. Set these PostgreSQL environment variables when connecting locally as well.

## Ordering and convergence

Marketplace child annotations give this local order:

```text
namespaces (0) → secret-store (1) → MongoDB + Meilisearch (2) → marketplace (3) → Kafka (4)
```

Waves order submission of child Applications only; they do not wait for each child to become Healthy and do not order the marketplace root against `platform`. Children converge asynchronously: unlimited retries with backoff (capped at five minutes) and `SkipDryRunOnMissingResource` absorb CRDs, storage, and secret stores that platform Applications have not finished installing. The only manual prerequisite is `ecommerce/doppler-token`.

Useful checks:

```bash
kubectl get applications -n argo-cd
kubectl get storageclass
kubectl get secretstore doppler -n ecommerce
kubectl get externalsecrets -n ecommerce
kubectl get crd | grep -E 'cnpg|strimzi|mongodb|external-secrets'
kubectl get svc vmagent-vmks vtagent-vmks -n observability
```

A child stuck retrying shows its last sync error in the environment's Argo CD UI. A missing or invalid token shows up on `prod-secret-store` first.

## Repository layout

```text
infra/argocd/
├── app-of-apps/
│   ├── values.yaml
│   └── templates/
│       ├── applications.tpl
│       └── namespaces.tpl
├── project.yaml
└── prod/root.yaml
```

See [ci.md](ci.md) for image publication, [networking.md](networking.md) for networking, and [observability.md](observability.md) for the telemetry contract. Platform details live in `talos-proxmox`'s [GitOps architecture](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/architecture/gitops.md) and [cluster access](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/operations/cluster-access.md) guides.
