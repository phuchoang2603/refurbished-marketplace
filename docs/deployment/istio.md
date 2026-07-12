# Istio mesh and ingress

Staging installs Istio ambient mode through four wrapper charts under `infra/charts/operators/istio/`, pinned to official Istio `1.30.2`. Marketplace browser traffic enters through Cloudflare Tunnel → an Istio Gateway API edge (`gatewayClassName: istio`).

## Components

| Wrapper   | Upstream chart                      | Argo CD Application  | Sync wave | Namespace      |
| --------- | ----------------------------------- | -------------------- | --------- | -------------- |
| `base`    | `istio/base`                        | `staging-istio-base` | 0         | `istio-system` |
| `istiod`  | `istio/istiod` (`profile=ambient`)  | `staging-istiod`     | 1         | `istio-system` |
| `cni`     | `istio/cni` (`profile=ambient`)     | `staging-istio-cni`  | 1         | `istio-system` |
| `ztunnel` | `istio/ztunnel` (`profile=ambient`) | `staging-ztunnel`    | 2         | `istio-system` |

Marketplace enrollment and Kafka follow at waves **3** and **4**.

## Marketplace enrollment (ambient + waypoint)

Staging enables ambient + waypoint via Helm values on `staging-refurbished-marketplace`:

```yaml
mesh:
  ambient:
    enabled: true
  waypoint:
    enabled: true
```

That labels the `ecommerce` namespace with `istio.io/dataplane-mode=ambient` and creates an `istio-waypoint` Gateway for east-west L7 telemetry (HBONE on port 15008).

Kafka/Connect/UI live in the separate `kafka` namespace (not ambient-enrolled) so Strimzi TLS and Debezium are not intercepted by ztunnel/waypoint. Marketplace services reach brokers at `ecommerce-kafka-cluster-kafka-bootstrap.kafka.svc:9092`.

**Tilt defaults keep `mesh.ambient.enabled: false`** so local pods are not redirected into ztunnel.

## Edge ingress (Gateway API + Cloudflare Tunnel)

Staging also enables north-south ingress from the marketplace chart:

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
```

Services are split by **subdomain** (Host-based `HTTPRoute`s), not by path under one host:

| Hostname             | Backend                     |
| -------------------- | --------------------------- |
| `shop.phuchoang.sbs` | `web`                       |
| `pay.phuchoang.sbs`  | `payment-gateway-simulator` |

| Resource             | `gatewayClassName` | Role                  |
| -------------------- | ------------------ | --------------------- |
| `ecommerce-waypoint` | `istio-waypoint`   | East-west L7 (mesh)   |
| `ecommerce-ingress`  | `istio`            | North-south HTTP edge |

Chart defaults keep `ingress.enabled: false` so Tilt continues to use port-forwards (`8080` → web, `8097` → simulator).

### Traffic path

| Layer                                         | Protocol                                   |
| --------------------------------------------- | ------------------------------------------ |
| Browser → Cloudflare                          | HTTPS (Cloudflare-managed cert)            |
| Cloudflare Tunnel → Istio Gateway             | HTTP to the Gateway Service / LoadBalancer |
| Gateway → `web` / `payment-gateway-simulator` | cluster HTTP (Host-based `HTTPRoute`)      |

No marketplace TLS Secret is required on the Istio Gateway for this path.

### Wire Cloudflare Tunnel (Proxmox LXC)

Keep `cloudflared` on the Proxmox LXC (not in-cluster). After the Gateway is Accepted:

1. Sync staging and wait for the ingress `Gateway` to become Accepted.
2. Find the LoadBalancer origin Istio created for the Gateway:

```bash
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce | grep ecommerce-ingress
```

3. On the LXC, configure Cloudflare Tunnel Public Hostnames:
   - `shop.phuchoang.sbs` → `http://<gateway-LB-IP>:80`
   - `pay.phuchoang.sbs` → `http://<gateway-LB-IP>:80`
4. Confirm those hostnames match the Helm `ingress.*Hostname` values exactly.
5. Open `https://shop.phuchoang.sbs` and exercise product → cart → checkout → payment (simulator redirect uses `HOSTED_PAYMENT_BASE_URL`).

## Protocol-aware Service ports

Marketplace Services render port names from `services.<name>.protocol`:

- HTTP: `web`, `payment-gateway-simulator`
- gRPC: `users`, `products`, `orders`, `cart`, `payment`

## Rollback

### Disable ingress only

1. Set `ingress.enabled: false` on the staging marketplace Application (or remove the `ingress:` block), restore a non-public `HOSTED_PAYMENT_BASE_URL` if needed, then sync.
2. Remove or repoint Cloudflare Tunnel Public Hostnames for the marketplace hosts.
3. Local Tilt port-forwards remain available with chart defaults.

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
kubectl get svc -n ecommerce | grep ecommerce-ingress
kubectl get pods -n kafka
```

Istio metrics are scraped by `staging-observability` (`istioScrapes` VMPodScrapes). In Grafana Explore (VictoriaMetrics):

```promql
sum by (source_app, destination_app) (rate(istio_tcp_connections_opened_total[5m]))
sum by (destination_app, request_protocol) (rate(istio_requests_total[5m]))
```

gRPC backends should show `request_protocol="grpc"` when L7 waypoint classification applies. Gateway entry can be confirmed via the ingress Gateway Service endpoints and related `istio_requests_total` series after browser traffic.

## Production

Production Istio Applications, marketplace mesh enrollment, and ingress enablement are intentionally omitted until staging is verified. Chart defaults keep `ingress.enabled: false`.
