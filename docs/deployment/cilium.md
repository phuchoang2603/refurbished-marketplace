# Cilium, Gateway API, and ingress

Cilium is installed by the **talos-proxmox** platform OpenTofu root before Argo CD, not by an Argo Application in this repo. This repo must not duplicate `apps/components/cilium` values (a second copy would drift and can wipe mesh flags on `helm upgrade`).

Already on the cluster (from talos-proxmox):

- CNI, kube-proxy replacement, L2 announcements, Gateway API CRDs
- WireGuard encryption, Envoy L7 proxy, cluster name/id as set in that repo
- L2 IP pools (`cilium-network` Application) and platform Gateways such as the Argo CD UI in `argo-cd`; Hubble UI is not required and may be absent

This repo only consumes that dataplane: marketplace `Gateway`/`HTTPRoute` (`gatewayClassName: cilium`) and a `TunnelBinding` for the platform's Cloudflare tunnel. Do not add WireGuard/Envoy/ClusterMesh values here.

Marketplace browser traffic: Cloudflare Tunnel → Cilium Gateway API.

Cilium 1.18 Gateway Services are `LoadBalancer`. `cloudflared` reaches the Gateway Service through in-cluster DNS (not the L2 VIP):

`http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`

East–west traffic is ordinary ClusterIP plus CiliumNetworkPolicy (allow-lists and optional required mTLS). Application traces and RED metrics are OTLP exports to the node-local `otel-agent` in `observability` (egress, so no CNP exception is needed). Hubble is not part of the observe path; request/error/duration SLIs are application OTEL metrics queried in HyperDX. See [observability.md](observability.md).

SPIRE for Cilium mutual auth is cluster Helm in **talos-proxmox** (`authentication.mutual.spire`). This chart sets `authentication.mode: required` on enrolled hops. Do not helm-upgrade Cilium from this repo.

## Allowed callers

Ingress policies select marketplace app pods only. Egress is unrestricted so CNPG, Valkey localhost, Kafka TLS, and OTLP keep working. Migration Jobs and CNPG Clusters are not selected.

| Destination                 | Port  | Allowed ingress                                                                                                   | mTLS (`mode: required`)  |
| --------------------------- | ----- | ----------------------------------------------------------------------------------------------------------------- | ------------------------ |
| `web`                       | 8080  | Cilium Gateway (`fromEntities: ingress`), kubelet (`host`), `payment-gateway-simulator` (hosted-payment callback) | Simulator → web only     |
| `users`                     | 9091  | `web`, kubelet                                                                                                    | web → users              |
| `products`                  | 9092  | `web`, kubelet                                                                                                    | web → products           |
| `catalog-mongodb`           | 27017 | `products`, kubelet, Kafka Connect (`kafka` ns, `strimzi.io/kind=KafkaConnect`)                                   | no                       |
| `catalog-meilisearch`       | 7700  | `search`, kubelet                                                                                                 | no                       |
| `search`                    | 9098  | `web`, kubelet                                                                                                    | web → search             |
| `inventory`                 | 9097  | `web`, kubelet                                                                                                    | web → inventory          |
| `orders`                    | 9093  | `web`, kubelet                                                                                                    | web → orders             |
| `cart`                      | 9094  | `web`, kubelet                                                                                                    | web → cart               |
| `payment`                   | 9096  | `web`, kubelet                                                                                                    | web → payment            |
| `payment-gateway-simulator` | 8097  | Cilium Gateway, kubelet                                                                                           | no (browser via Gateway) |

Inventory is a separate gRPC service with its own CNPG cluster. Checkout holds stock with `ReserveStock` from `web`; products is not an allowed inventory caller. Kafka consumers are those same service pods talking to namespace `kafka` (Strimzi TLS, not mesh mTLS). Cart → Valkey is `127.0.0.1` on the pod. Init/migrate containers talk to remaining `*-db-rw:5432` without a CNP on CNPG. Mongo 27017 is allow-listed for products, kubelet, and Kafka Connect only; that hop does not use SPIRE. Meilisearch 7700 is allow-listed for search and kubelet only; that hop does not use SPIRE.

Chart knobs (`infra/charts/refurbished-marketplace/values.yaml`):

| Value                   | Effect                                                                                                    |
| ----------------------- | --------------------------------------------------------------------------------------------------------- |
| `meshPolicy.enabled`    | Render CNPs. `false` restores post-Istio ClusterIP (no allow-list).                                       |
| `meshPolicy.enforce`    | `true` sets `enableDefaultDeny.ingress: true` on CNPs (unknown callers dropped). `false` is observe-only. |
| `meshPolicy.mutualAuth` | `false` drops `authentication.mode: required` (identity allow-list without SPIRE handshake).              |

## Gateway timeouts and outlier detection

Shop and pay HTTPRoutes set `timeouts.request` / `timeouts.backendRequest` (defaults 30s / 25s). **No HTTPRoute retries** and no gRPC retry interceptors — checkout POST and hosted-payment callbacks are not idempotent at this layer ([#35](https://github.com/phuchoang2603/refurbished-marketplace/issues/35) is required before adding retries).

Cilium Gateway applies Envoy outlier detection on backend clusters by default. This repo does not add `CiliumEnvoyConfig`.

## Policy verification

After CNPs sync with `meshPolicy.enforce: true`:

```bash
kubectl get ciliumnetworkpolicy -n ecommerce
kubectl -n kube-system -c cilium-agent logs -l k8s-app=cilium --tail=200 \
  | grep -E 'Policy is requiring authentication|Successfully authenticated|Policy denied'

# Unknown caller — expect connection failure (not HTTP 200):
kubectl -n ecommerce run policy-probe --restart=Never --image=busybox:1.36 \
  --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":65534,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"policy-probe","image":"busybox:1.36","securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true,"runAsUser":65534,"seccompProfile":{"type":"RuntimeDefault"}},"command":["wget","-qO-","--timeout=3","http://users:9091"]}]}}' \
  --command -- wget -qO- --timeout=3 http://users:9091
kubectl -n ecommerce wait --for=jsonpath='{.status.phase}'=Failed pod/policy-probe --timeout=60s
kubectl -n ecommerce logs policy-probe   # expect wget error / non-zero exit
kubectl -n ecommerce delete pod policy-probe

# Allowed path — shop browse and checkout on talos-dev:
curl -fsS -o /dev/null -w '%{http_code}\n' https://shop-dev.phuchoang.sbs/
curl -fsS -o /dev/null -w '%{http_code}\n' https://shop-dev.phuchoang.sbs/products
# Complete one checkout in the browser (login → cart → pay simulator → return).
```

## Mesh policy rollback

Git + Argo sync, no app code change:

1. `meshPolicy.mutualAuth: false` — keep allow-lists, disable required mTLS.
2. `meshPolicy.enforce: false` — observe / open default-deny.
3. `meshPolicy.enabled: false` — remove CNPs entirely.

Then sync the marketplace Application. The `TunnelBinding` and Cloudflare hostnames stay put.

## GitOps

`dev-root` / `prod-root` (`infra/argocd/dev/root.yaml`, `infra/argocd/prod/root.yaml`) run on each environment's own Argo CD and render [`infra/argocd/app-of-apps`](../../infra/argocd/app-of-apps/). Marketplace enrollment has no Istio labels. Kafka stays in namespace `kafka`.

## Edge

| Env  | Hostname                 | Backend                     |
| ---- | ------------------------ | --------------------------- |
| dev  | `shop-dev.phuchoang.sbs` | `web`                       |
| dev  | `pay-dev.phuchoang.sbs`  | `payment-gateway-simulator` |
| prod | `shop.phuchoang.sbs`     | `web`                       |
| prod | `pay.phuchoang.sbs`      | `payment-gateway-simulator` |

HTTPRoutes set `X-Forwarded-Proto: https` and `X-Forwarded-Host` so hosted-payment callbacks are not rewritten to HTTP (POST → Cloudflare 301 → GET → 405).

The chart renders `TunnelBinding/ecommerce-ingress` (`ingress.tunnel`) with both hostnames as subjects of the `cilium-gateway-ecommerce-ingress` Service. The `talos-proxmox` Cloudflare operator adds them to the environment's `ClusterTunnel/talos-proxmox`, restarts `cloudflared`, and creates each hostname's proxied CNAME and `_managed.<hostname>` ownership TXT record. Deleting the binding, or disabling `ingress.tunnel`, removes the records. Nothing is configured in the Cloudflare dashboard.

The operator refuses a hostname that already has a DNS record it does not own. Delete such a record in Cloudflare and the binding converges on its next retry.

TLS terminates at Cloudflare. No marketplace TLS Secret on the Gateway. Do not reuse the platform Argo CD or HyperDX Gateways for shop/pay.

```bash
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl get tunnelbinding -n ecommerce
kubectl describe tunnelbinding ecommerce-ingress -n ecommerce   # DNS and config events
kubectl get pods -n cloudflare-operator-system
```

## Rollback

Disable marketplace `ingress.tunnel.enabled` (removes the public hostnames) or `ingress.enabled` and sync. Cilium itself stays with talos-proxmox.
