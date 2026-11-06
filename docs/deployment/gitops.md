# GitOps deployment (Argo CD)

Proxmox **dev** and GCP **prod** each run their own Argo CD in namespace `argo-cd`, installed by the [`talos-proxmox`](https://github.com/phuchoang2603/talos-proxmox) platform OpenTofu root. Argo CD manages its local cluster and hosts two independent roots:

1. `platform` from `talos-proxmox` installs shared operators, the Doppler `ClusterSecretStore`, storage, observability, and the Cloudflare operator with its `ClusterTunnel`.
2. `dev-root` or `prod-root` from this repository deploys marketplace-owned resources that consume those platform APIs.

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

Bootstrap once per environment. Apply the Doppler token as described in [secrets.md](../development/secrets.md), then the project and root with the same kubeconfig:

```bash
kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply --server-side -f infra/argocd/prod/root.yaml
```

## Environment configuration

| Layer                      | Dev                                | Prod                           |
| -------------------------- | ---------------------------------- | ------------------------------ |
| Argo CD                    | `https://argocd-dev.phuchoang.sbs` | `https://argocd.phuchoang.sbs` |
| Platform kubeconfig        | `talos-proxmox/dev`                | `talos-proxmox/prod`           |
| Marketplace root           | `dev-root`                         | `prod-root`                    |
| Marketplace Doppler config | `dev`                              | `prd`                          |
| Root Git ref               | `deploy/dev`                       | `deploy/prod`                  |
| Shop                       | `shop-dev.phuchoang.sbs`           | `shop.phuchoang.sbs`           |

Apply `infra/argocd/project.yaml` and the matching `infra/argocd/{dev,prod}/root.yaml` once using that environment's kubeconfig. Each root watches its deployment ref. After image publication, CI updates that ref with the immutable chart source revision and per-service commit tags inherited by its children. Roots select the matching `values-dev.yaml` or `values-prod.yaml` overlays. Subsequent releases need no Kubernetes API access from CI. A root may initially report an unavailable ref until the first release publishes it.

The marketplace chart's `TunnelBinding` publishes shop through the platform's `ClusterTunnel/talos-proxmox`; the Cloudflare operator creates its DNS records. The origin is `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`. See [networking.md](networking.md#edge).

Production Checkout uses `PAYMENT_GATEWAY_URL=http://edge.payment-gateway.svc.cluster.local:8080`, `CHECKOUT_RETURN_BASE_URL=https://shop.phuchoang.sbs`, and the two keys from `checkout-payment-gateway`; see [secrets.md](../development/secrets.md). The API key identifies the integration at Edge. Chart defaults leave the gateway disabled and refuse checkout submission. The marketplace no longer renders Payment, its database, or the hosted-page simulator, and publishes no `pay` / `pay-dev` hostname.

Dev intentionally leaves gateway integration disabled: browsing, accounts, cart and search are in scope, while checkout submissions return unavailable without creating orders or reservations. Production checkout verification requires the separate gateway deployment. Provision the gateway integration and matching credentials before deploying this release; see [the rollout gate](../order-placement.md#rollout-gate). Local Helm renders do not satisfy that gate.

Dev deploys the web service as an Argo Rollout with two replicas and Gateway API traffic weights of 5%, 25%, 50%, and 100%. Each step checks the candidate revision's catalog request volume, metric freshness, server error rate (at most 1%), and p95 latency (at most 500 ms) in VictoriaMetrics. Missing or invalid metrics fail the analysis. Supply catalog traffic during updates; an idle candidate cannot pass the minimum of 20 requests per minute. Production keeps its Deployment unless `rollout.enabled` is explicitly enabled. The platform Argo Rollouts controller requires Gateway API permissions in `ecommerce`; Argo CD ignores only the selectors and route weights that Rollouts manages.

Run the in-cluster storefront benchmark from the project devenv. Seed once, then reuse the dataset for each load level:

```bash
python3 benchmarks/run.py seed --kubeconfig "$HOME/.kube/talos-dev.yaml" \
  --dataset /tmp/marketplace-dataset.json --output /tmp/marketplace-seed
python3 benchmarks/run.py load --kubeconfig "$HOME/.kube/talos-dev.yaml" \
  --dataset /tmp/marketplace-dataset.json --rate 50 --duration 2m \
  --output /tmp/marketplace-50rps
python3 benchmarks/run.py load --workload inventory \
  --kubeconfig "$HOME/.kube/talos-dev.yaml" \
  --dataset /tmp/marketplace-dataset.json --rate 100 --duration 2m \
  --output /tmp/marketplace-inventory-100rps
```

The seed creates 1,000 synthetic listings through the HTTP application and waits for the first and last listing to appear in the Kafka-backed search projection. The load is 45% catalog browse, 35% text search, and 20% product detail through the internal Istio Gateway. Catalog responses must contain 20 cards by default; use `--page-size 100` when comparing the old page size or `--page-size 0` only for mixed-version rollout traffic. Set `--vus` and `--max-vus` to preallocate enough clients for a load sweep. Local results include raw k6 logs, JSON metrics, parameters, and timestamps. All output under `benchmarks/results/` is ignored by Git. Recorded results are summarized in [the dev benchmark report](dev-benchmarks.md). This measures server latency within the dev cluster; it excludes Cloudflare, browser rendering, and checkout/payment processing. Jobs expire after one day; synthetic accounts and products remain in the dev database.

The inventory workload measures `InventoryService/GetStock` through the Istio waypoint against the same product IDs, checks each returned stock quantity, and uses a 100 ms p95 threshold. Run workloads sequentially for isolated measurements. Dev grants each Go service a 1-CPU limit and Meilisearch a 2-CPU / 512-MiB limit. PostgreSQL application pools retain up to 16 connections per pod, leaving room within each database's connection budget for rollout overlap, migrations, and CDC.

## Marketplace ownership

| Application               | Marketplace-owned resources                                                                                                                              | Namespace                    |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| root (`prod-root`)        | `ecommerce` Namespace with `istio.io/dataplane-mode: ambient`                                                                                            | cluster                      |
| `secret-store`            | `SecretStore/doppler`, the single owner shared by every marketplace ExternalSecret                                                                       | `ecommerce`                  |
| `meilisearch`             | Meilisearch workload/PVC, credentials                                                                                                                    | `ecommerce`                  |
| `refurbished-marketplace` | CNPG Clusters, ExternalSecrets, migrations, services, Gateways/HTTPRoutes, TunnelBinding, PeerAuthentication, DestinationRules, gateway webhook policies | `ecommerce`                  |
| `kafka`                   | Kafka/NodePool, topics, Connect/connectors, secret-reader RBAC, UI, PeerAuthentication                                                                   | `ecommerce`                  |
| `cloudflare-tunnel`       | platform-owned in `talos-proxmox`                                                                                                                        | `cloudflare-operator-system` |

The `platform` root owns the operators, CRDs, the Cloudflare tunnel and its cloudflared, and the telemetry pipeline (VMAgent, VTAgent, VLAgent, Victoria stores and Grafana on prod). Marketplace workloads only export OTLP/HTTP to VMAgent and VTAgent; see [observability.md](observability.md). The marketplace `secret-store` Application owns `SecretStore/doppler` in `ecommerce`; no other chart renders it. Cilium, Istio, Gateway API CRDs, and storage are also owned by `talos-proxmox`. Child Applications do not use `CreateNamespace`; only the root creates and labels the marketplace namespaces.

CNPG PostgreSQL is the products catalog source of truth; Meilisearch is its storefront projection. The catalog cutover starts empty. Products uses the same Goose migration init container and separate owner/runtime roles as the other PostgreSQL services.

Cart and Checkout use separate persistent `Redis` resources (`cart-redis` and `checkout-redis`) from the platform's OT-CONTAINER-KIT Redis operator. Both use AOF and `noeviction`; the checkout projection also recovers its account facts and checkpoint from the retained Users topic. Deploy the Users migration and connector before enabling the new checkout image, and wait for account replay before the checkout smoke test.

Fresh installations start with empty cart and checkout Redis instances; no cart-data copy or legacy sidecar cutover is required. The release workflow publishes the chart revision and per-service commit tags to the environment deployment ref after affected images build.

Each PostgreSQL service image supports `migrate`, using SQL embedded in the binary and Goose's Go SDK with a PostgreSQL advisory lock. A `migrate` init container runs before each application container; a failure blocks startup and Kubernetes retries it. There are no migration Jobs or separate migrator images. Existing `goose_db_version` history is preserved.

Migration containers use the database owner Secret (`<service>-app`); application containers use restricted `<service>_runtime` roles and separate `<service>-runtime` Secrets. CNPG manages both roles. After applying schema changes, the migration command grants runtime DML access to application tables and sequences and default privileges for future tables, while withholding access to `goose_db_version` and schema creation. Both containers use `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, and `PGSSLMODE`; passwords are read directly from Secrets without connection-string interpolation. Provision the five `*_RUNTIME_PASSWORD` Doppler keys before deploying, and build and pin the new service images before enabling the init containers.

Kafka, Connect, connectors, topics, and their RBAC all run in `ecommerce`; the root no longer renders a `kafka` Namespace. This namespace change starts a fresh Kafka cluster and discards the old Kafka data when the old namespace is pruned. Pause marketplace consumers and the old connectors before that cutover, remove their PostgreSQL replication slots (`debezium_<service>_slot`) once those connectors stop, then let the new PostgreSQL connectors snapshot again. Reset the Checkout account projection checkpoint in Redis with the new Kafka cluster to avoid using offsets from the old Users topic.

## Ordering and convergence

Marketplace child annotations give this local order:

```text
namespaces (0) → secret-store (1) → Meilisearch (2) → marketplace (3) → Kafka (4)
```

Waves order submission of child Applications only; they do not wait for each child to become Healthy and do not order the marketplace root against `platform`. Children converge asynchronously: unlimited retries with backoff (capped at five minutes) and `SkipDryRunOnMissingResource` absorb CRDs, storage, and secret stores that platform Applications have not finished installing. The one-time bootstrap prerequisite is `ecommerce/doppler-token`.

Useful checks:

```bash
kubectl get applications -n argo-cd
kubectl get storageclass
kubectl get secretstore doppler -n ecommerce
kubectl get externalsecrets -n ecommerce
kubectl get crd | grep -E 'cnpg|strimzi|external-secrets'
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
├── dev/root.yaml
└── prod/root.yaml
```

See [ci.md](ci.md) for image publication, [networking.md](networking.md) for networking, and [observability.md](observability.md) for the telemetry contract. Platform details live in `talos-proxmox`'s [GitOps architecture](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/architecture/gitops.md) and [cluster access](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/operations/cluster-access.md) guides.
