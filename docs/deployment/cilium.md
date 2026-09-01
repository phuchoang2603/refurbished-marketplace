# Cilium, Gateway API, and ingress

Cilium is cluster bootstrap from **talos-proxmox**, not an Argo Application in this repo. Empty-cluster CNI still needs that Helm install; this repo must not duplicate `apps/values/cilium.yaml` (a second copy would drift and can wipe mesh flags on `helm upgrade`).

Already on the cluster (Cilium Helm from talos-proxmox):

- CNI, kube-proxy replacement, L2 announcements, Gateway API
- WireGuard encryption, Envoy L7 proxy, cluster name/id as set in that repo
- L2 IP pool + platform Gateways in `cilium-ingress` (Longhorn, Argo CD; Hubble UI is not required and may be absent)

This repo only consumes that dataplane: marketplace and Grafana `Gateway`/`HTTPRoute` (`gatewayClassName: cilium`) and Cloudflare origin DNS. Do not add WireGuard/Envoy/ClusterMesh values here.

Marketplace browser traffic: Cloudflare Tunnel → Cilium Gateway API.

Cilium 1.18 Gateway Services are `LoadBalancer`. `cloudflared` uses in-cluster DNS (not the L2 VIP):

`http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`

East–west traffic is ordinary ClusterIP plus CiliumNetworkPolicy (allow-lists and optional required mTLS). Application traces stay OTEL → VictoriaTraces. Hubble is not part of the observe path; request/error/duration SLIs are follow-on issue [#43](https://github.com/phuchoang2603/refurbished-marketplace/issues/43).

SPIRE for Cilium mutual auth is cluster Helm in **talos-proxmox** (`authentication.mutual.spire`). This chart sets `authentication.mode: required` on enrolled hops. Do not helm-upgrade Cilium from this repo.

## Allowed callers

Ingress policies select marketplace app pods only. Egress is unrestricted so CNPG, Valkey localhost, Kafka TLS, and OTLP keep working. Migration Jobs and CNPG Clusters are not selected.

| Destination                 | Port | Allowed ingress                                                                                                   | mTLS (`mode: required`)  |
| --------------------------- | ---- | ----------------------------------------------------------------------------------------------------------------- | ------------------------ |
| `web`                       | 8080 | Cilium Gateway (`fromEntities: ingress`), kubelet (`host`), `payment-gateway-simulator` (hosted-payment callback) | Simulator → web only     |
| `users`                     | 9091 | `web`, kubelet                                                                                                    | web → users              |
| `products`                  | 9092 | `web`, kubelet                                                                                                    | web → products           |
| `orders`                    | 9093 | `web`, kubelet                                                                                                    | web → orders             |
| `cart`                      | 9094 | `web`, kubelet                                                                                                    | web → cart               |
| `payment`                   | 9096 | `web`, kubelet                                                                                                    | web → payment            |
| `payment-gateway-simulator` | 8097 | Cilium Gateway, kubelet                                                                                           | no (browser via Gateway) |

Inventory lives in `products` (catalog). Kafka consumers are those same service pods talking to namespace `kafka` (Strimzi TLS, not mesh mTLS). Cart → Valkey is `127.0.0.1` on the pod. Init/migrate containers talk to `*-db-rw:5432` without a CNP on CNPG.

Chart knobs (`infra/charts/refurbished-marketplace/values.yaml`):

| Value                   | Effect                                                                                       |
| ----------------------- | -------------------------------------------------------------------------------------------- |
| `meshPolicy.enabled`    | Render CNPs. `false` restores post-Istio ClusterIP (no allow-list).                          |
| `meshPolicy.enforce`    | `false` keeps allow-lists but sets `enableDefaultDeny.ingress: false` (observe).             |
| `meshPolicy.mutualAuth` | `false` drops `authentication.mode: required` (identity allow-list without SPIRE handshake). |

## Timeouts, retries, circuit breaking

Shop and pay HTTPRoutes set `timeouts.request` / `timeouts.backendRequest` (defaults 30s / 25s). **No HTTPRoute retries** and the Go gRPC clients have no retry interceptor: `CreateOrder` and hosted-payment callbacks must not be retried at the dataplane. Cilium Gateway already applies Envoy outlier detection on Gateway backend clusters; this repo does not add `CiliumEnvoyConfig` (it would intercept ClusterIP and is easy to get wrong).

## Policy verification

Hubble is optional/off. After CNPs sync:

```bash
kubectl get ciliumnetworkpolicy -n ecommerce
kubectl -n kube-system -c cilium-agent logs -l k8s-app=cilium --tail=200 | grep -E "Policy is requiring authentication|Successfully authenticated|Denied"
# Unknown caller (expect drop when enforce is true):
kubectl -n ecommerce run policy-probe --restart=Never --image=busybox --command -- wget -qO- --timeout=3 http://users:9091
kubectl -n ecommerce delete pod policy-probe
```

Checkout through `shop-dev` / `shop` should still succeed for allowed callers.

## Mesh policy rollback

Git + Argo sync, no app code change:

1. `meshPolicy.mutualAuth: false` — keep allow-lists, disable required mTLS.
2. `meshPolicy.enforce: false` — observe / open default-deny.
3. `meshPolicy.enabled: false` — remove CNPs entirely.

Then sync the marketplace Application. Cloudflare hostnames stay put.

## GitOps

`dev-root` / `prod-root` (`infra/argocd/dev/root.yaml`, `infra/argocd/prod/root.yaml`) render [`infra/argocd/app-of-apps`](../../infra/argocd/app-of-apps/). Marketplace enrollment has no Istio labels. Kafka stays in namespace `kafka`.

## Edge

| Env  | Hostname                    | Backend                     |
| ---- | --------------------------- | --------------------------- |
| dev  | `shop-dev.phuchoang.sbs`    | `web`                       |
| dev  | `pay-dev.phuchoang.sbs`     | `payment-gateway-simulator` |
| dev  | `grafana-dev.phuchoang.sbs` | `observability-grafana`     |
| prod | `shop.phuchoang.sbs`        | `web`                       |
| prod | `pay.phuchoang.sbs`         | `payment-gateway-simulator` |
| prod | `grafana.phuchoang.sbs`     | `observability-grafana`     |

HTTPRoutes set `X-Forwarded-Proto: https` and `X-Forwarded-Host` so hosted-payment callbacks are not rewritten to HTTP (POST → Cloudflare 301 → GET → 405).

Cloudflare Zero Trust Public Hostnames (not in Git):

- `shop-dev.phuchoang.sbs` / `pay-dev.phuchoang.sbs` (dev) → `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`
- `grafana-dev.phuchoang.sbs` (dev) → `http://cilium-gateway-grafana.monitoring.svc.cluster.local:80`
- `shop.phuchoang.sbs` / `pay.phuchoang.sbs` (prod) → same origin DNS on the prod cluster
- `grafana.phuchoang.sbs` (prod) → `http://cilium-gateway-grafana.monitoring.svc.cluster.local:80`

TLS terminates at Cloudflare. No marketplace TLS Secret on the Gateway. Do not reuse the `cilium-ingress` Longhorn/Argo Gateways for shop/pay.

```bash
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl get pods -n cloudflare-tunnel
```

## Rollback

Disable marketplace `ingress.enabled` and sync, or repoint Cloudflare hostnames. Cilium itself stays with talos-proxmox.
