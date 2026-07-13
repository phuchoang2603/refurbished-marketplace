# Istio mesh and ingress

Staging installs Istio ambient mode through four wrapper charts under `infra/charts/operators/istio/`, pinned to official Istio `1.30.2`. Marketplace browser traffic enters through in-cluster Cloudflare Tunnel (`cloudflared`) → an Istio Gateway API edge (`gatewayClassName: istio`).

## Components

| Wrapper   | Upstream chart                      | Argo CD Application  | Sync wave | Namespace      |
| --------- | ----------------------------------- | -------------------- | --------- | -------------- |
| `base`    | `istio/base`                        | `staging-istio-base` | 0         | `istio-system` |
| `istiod`  | `istio/istiod` (`profile=ambient`)  | `staging-istiod`     | 1         | `istio-system` |
| `cni`     | `istio/cni` (`profile=ambient`)     | `staging-istio-cni`  | 1         | `istio-system` |
| `ztunnel` | `istio/ztunnel` (`profile=ambient`) | `staging-ztunnel`    | 2         | `istio-system` |

Marketplace enrollment and Kafka follow at waves **3** and **4**. The Cloudflare Tunnel connector syncs at wave **5** (after the marketplace ingress Gateway).

## Marketplace enrollment (ambient + waypoint)

Staging enables ambient + waypoint via chart `values.yaml`, with production hostnames/GHCR from `values-staging.yaml` on `staging-refurbished-marketplace`:

```yaml
# values-staging.yaml (excerpt)
ingress:
  enabled: true
  webHostname: shop.phuchoang.sbs
  simulatorHostname: pay.phuchoang.sbs
```

That labels the `ecommerce` namespace with `istio.io/dataplane-mode=ambient` and creates an `istio-waypoint` Gateway for east-west L7 telemetry (HBONE on port 15008).

Kafka/Connect/UI live in the separate `kafka` namespace (not ambient-enrolled) so Strimzi TLS and Debezium are not intercepted by ztunnel/waypoint. Marketplace services reach brokers at `ecommerce-kafka-cluster-kafka-bootstrap.kafka.svc:9092`.

Chart `values.yaml` enables ambient, waypoint, and ingress for local Colima (`.dev` hosts + Cloudflare Tunnel). Staging overlays `values-staging.yaml` (production hostnames, GHCR). Istio CNI defaults to `global.platform=k3s` in `values.yaml`; staging uses `values-staging.yaml` for RKE2 paths.

## Edge ingress (Gateway API + Cloudflare Tunnel)

Staging enables north-south ingress from the marketplace chart:

```yaml
ingress:
  enabled: true
  name: ecommerce-ingress
  port: 80
  webHostname: shop.phuchoang.sbs
  simulatorHostname: pay.phuchoang.sbs
services:
  web:
    env:
      HOSTED_PAYMENT_BASE_URL: https://pay.phuchoang.sbs
      PUBLIC_BASE_URL: https://shop.phuchoang.sbs
      HOSTED_PAYMENT_CALLBACK_BASE_URL: http://web:8080
```

The web `HTTPRoute` sets `X-Forwarded-Proto: https` because TLS terminates at Cloudflare and the origin sees plain HTTP. Without that (or `PUBLIC_BASE_URL`), hosted-payment callback URLs are built as `http://…`, Cloudflare’s HTTPS redirect turns the simulator’s POST into a GET, and `/callbacks/hosted-payment` returns 405.

Services are split by **subdomain** (Host-based `HTTPRoute`s), not by path under one host:

| Hostname             | Backend                     |
| -------------------- | --------------------------- |
| `shop.phuchoang.sbs` | `web`                       |
| `pay.phuchoang.sbs`  | `payment-gateway-simulator` |

| Resource             | `gatewayClassName` | Role                                      |
| -------------------- | ------------------ | ----------------------------------------- |
| `ecommerce-waypoint` | `istio-waypoint`   | East-west L7 (mesh)                       |
| `ecommerce-ingress`  | `istio`            | North-south HTTP edge (ClusterIP Service) |

Chart `values.yaml` enables ingress for local `.dev` hosts; staging `values-staging.yaml` sets production hostnames. Local and staging both reach the Gateway through Cloudflare Tunnel (no per-service port-forwards).

### Traffic path

| Layer                                         | Protocol                                                         |
| --------------------------------------------- | ---------------------------------------------------------------- |
| Browser → Cloudflare                          | HTTPS (Cloudflare-managed cert)                                  |
| Cloudflare → in-cluster `cloudflared`         | Tunnel connector in `cloudflare-tunnel`                          |
| `cloudflared` → Istio Gateway                 | HTTP to `ecommerce-ingress-istio.ecommerce.svc.cluster.local:80` |
| Gateway → `web` / `payment-gateway-simulator` | cluster HTTP (Host-based `HTTPRoute`)                            |

No marketplace TLS Secret is required on the Istio Gateway for this path.

### In-cluster cloudflared (Argo CD)

`staging-cloudflare-tunnel` deploys `infra/charts/cloudflare-tunnel` into namespace `cloudflare-tunnel`. The tunnel token is synced from Doppler key `CLOUDFLARE_TUNNEL_TOKEN` via External Secrets (not committed to Git).

1. Create a Cloudflare Zero Trust tunnel and copy the token into Doppler `prd` as `CLOUDFLARE_TUNNEL_TOKEN`.
2. Sync staging; wait for `ecommerce-ingress` Gateway and `cloudflared` pods to become Ready.
3. In Cloudflare Zero Trust → Public Hostnames, set:
   - `shop.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`
   - `pay.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`
4. Confirm those hostnames match the Helm `ingress.*Hostname` values exactly.
5. Open `https://shop.phuchoang.sbs` and exercise product → cart → checkout → payment.

```bash
kubectl get gateway,httproute -n ecommerce
kubectl get svc ecommerce-ingress-istio -n ecommerce
kubectl get pods -n cloudflare-tunnel
kubectl get externalsecret,secret -n cloudflare-tunnel
```

## Protocol-aware Service ports

Marketplace Services render port names from `services.<name>.protocol`:

- HTTP: `web`, `payment-gateway-simulator`
- gRPC: `users`, `products`, `orders`, `cart`, `payment`

## Rollback

### Disable ingress / tunnel

1. Set `ingress.enabled: false` on the staging marketplace Application (or remove the `ingress:` block), restore a non-public `HOSTED_PAYMENT_BASE_URL` if needed, then sync.
2. Optionally remove or disable `staging-cloudflare-tunnel`, and remove or repoint Cloudflare Public Hostnames.
3. Local access uses Cloudflare Tunnel to `shop.dev.phuchoang.sbs` / `pay.dev.phuchoang.sbs` (see [local-setup](../development/local-setup.md)).

### Disable mesh enrollment

1. Disable enrollment first — set `mesh.ambient.enabled: false` and `mesh.waypoint.enabled: false` on the staging marketplace Application (or remove those values), then sync.
2. Restart marketplace pods so they leave ambient redirection.
3. Only then prune/disable the Istio Applications (`staging-ztunnel`, `staging-istio-cni`, `staging-istiod`, `staging-istio-base`) if you need to remove the mesh platform.

Do not delete Istio while `ecommerce` is still labeled for ambient mode.

## Canal / NetworkPolicy note

Staging uses RKE2 Canal. Ambient HBONE uses TCP **15008**. Any allow-list NetworkPolicy that covers mesh traffic must permit that port.

## Verification

```bash
kubectl get pods -n istio-system
kubectl get daemonset -n istio-system
kubectl get ns ecommerce --show-labels
kubectl get gateway,httproute -n ecommerce
kubectl get svc ecommerce-ingress-istio -n ecommerce
kubectl get pods -n cloudflare-tunnel
kubectl get pods -n kafka
```

Istio metrics are scraped by `staging-observability` (`istioScrapes` VMPodScrapes). In Grafana Explore (VictoriaMetrics):

```promql
sum by (source_app, destination_app) (rate(istio_tcp_connections_opened_total[5m]))
sum by (destination_app, request_protocol) (rate(istio_requests_total[5m]))
```

gRPC backends should show `request_protocol="grpc"` when L7 waypoint classification applies. Gateway entry can be confirmed via the ingress Gateway Service and related `istio_requests_total` series after browser traffic.

## Production

Production Istio Applications, marketplace mesh enrollment, ingress enablement, and cloudflared are intentionally omitted until staging is verified.
