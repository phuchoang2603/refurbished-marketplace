# Networking, Istio ambient, and ingress

The **talos-proxmox** platform owns the cluster network. Cilium is the CNI and L2/LB IPAM provider; its Gateway controller, SPIRE authentication, and L7 proxy are disabled. Istio (base, istiod, CNI, ztunnel) supplies ambient mesh transport and the `istio` / `istio-waypoint` GatewayClasses. This repo installs neither and must not duplicate their Helm values.

This repo consumes that dataplane:

- The marketplace root renders the `ecommerce` and `kafka` Namespaces with `istio.io/dataplane-mode: ambient`. Child Applications never create or relabel them, and no platform namespace is enrolled.
- Each namespace has a `STRICT` `PeerAuthentication` (marketplace chart for `ecommerce`, Kafka chart for `kafka`). Kafka keeps its own broker TLS; ambient mTLS is transport protection, not a Kafka replacement.
- The marketplace chart renders the browser `Gateway`/`HTTPRoute`s (`gatewayClassName: istio`), a service waypoint, and per-service `DestinationRule`s.

No CiliumNetworkPolicy or per-service allow-list is rendered. Application traces and RED metrics are OTLP exports to the node-local `otel-agent` in `observability`; see [observability.md](observability.md).

## Edge

Browser traffic: Cloudflare Tunnel → Istio Gateway → `web` / `payment-gateway-simulator`.

| Env  | Hostname                 | Backend                     |
| ---- | ------------------------ | --------------------------- |
| dev  | `shop-dev.phuchoang.sbs` | `web`                       |
| dev  | `pay-dev.phuchoang.sbs`  | `payment-gateway-simulator` |
| prod | `shop.phuchoang.sbs`     | `web`                       |
| prod | `pay.phuchoang.sbs`      | `payment-gateway-simulator` |

Istio provisions `Service/ecommerce-ingress-istio` for `Gateway/ecommerce-ingress`. The Gateway's `networking.istio.io/service-type: ClusterIP` annotation keeps it off the Cilium L2 pool; the tunnel reaches it through in-cluster DNS:

`http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`

HTTPRoutes set `X-Forwarded-Proto: https` and `X-Forwarded-Host` so hosted-payment callbacks are not rewritten to HTTP (POST → Cloudflare 301 → GET → 405).

The chart renders `TunnelBinding/ecommerce-ingress` (`ingress.tunnel`) with both hostnames and an explicit `target` on that origin, because the generated Service also exposes the Istio status port. The `talos-proxmox` Cloudflare operator adds them to the environment's `ClusterTunnel/talos-proxmox`, restarts `cloudflared`, and creates each hostname's proxied CNAME and `_managed.<hostname>` ownership TXT record. Deleting the binding, or disabling `ingress.tunnel`, removes the records. Nothing is configured in the Cloudflare dashboard.

The operator refuses a hostname that already has a DNS record it does not own. Delete such a record in Cloudflare and the binding converges on its next retry.

TLS terminates at Cloudflare. No marketplace TLS Secret on the Gateway. Do not reuse the platform Argo CD or HyperDX Gateways for shop/pay.

## Waypoint and resilience

`Gateway/waypoint` (`istio-waypoint`, `istio.io/waypoint-for: service`) handles L7 traffic for the gRPC Services listed in `mesh.resilience.services`; those Services carry `istio.io/use-waypoint: waypoint`. `web`, the simulator, Kafka, MongoDB, Meilisearch, and CNPG stay on ztunnel only. The ingress Gateway is itself an Envoy that enforces the shop and pay route policies, and its only backends (`web`, the simulator) have no waypoint, so no Service sets `istio.io/ingress-use-waypoint`. Add that label only if a waypoint-enrolled Service becomes an ingress backend.

Each listed Service gets a `DestinationRule` with connection and pending-request caps and a retry budget. Outlier ejection is rendered only for Services with `replicas > 1`, so a sole endpoint is never ejected.

Retry rules depend on two platform prerequisites from `talos-proxmox`: mesh-wide default HTTP retries disabled (`meshConfig.defaultHttpRetryPolicy.attempts: 0`) and the Gateway API Experimental CRDs, which admit `HTTPRoute.rules[].retry`.

- **Ingress reads** (`ingress.reads`): shop and pay `GET`s retry twice on connect failures, refused streams, and 502/503, with a 9s per-attempt and 30s total budget.
- **Ingress writes** (`ingress.writes`): every other request, including checkout, cart, hosted-payment, and callback POSTs, has a 30s total / 25s backend budget and never retries.
- **Waypoint gRPC** (`mesh.resilience.grpcRetries`): each listed Service gets a `<name>-mesh` HTTPRoute attached at the waypoint. Exact unary method paths for lookups and `SubmitCheckout` retry twice with a 3s per-attempt and 10s total budget; an identical `SubmitCheckout` (same buyer, intent key, and payload) returns the original checkout. Istio retries connect failures, refused streams, and gRPC `UNAVAILABLE`/`CANCELLED` responses; it does not retry other gRPC statuses. The catch-all rule routes every other method, including cart and payment mutations, without retries or a timeout.

Proxy retries do not replace the checkout saga's inbox/outbox replay, deadlines, or compensation, and Istio never retries Kafka delivery.

| Value                              | Effect                                                                         |
| ---------------------------------- | ------------------------------------------------------------------------------ |
| `ingress.reads`                    | Method, timeouts, and retry for the shop and pay read rule.                    |
| `ingress.writes`                   | Timeouts for the unretried shop and pay catch-all rule.                        |
| `mesh.waypoint.enabled`            | Render the waypoint, its options ConfigMap, Service labels, and mesh policies. |
| `mesh.waypoint.replicas`           | Waypoint Deployment replicas; a PodDisruptionBudget is added above one.        |
| `mesh.resilience.services`         | gRPC Services routed through the waypoint.                                     |
| `mesh.resilience.connectionPool`   | DestinationRule connection and pending-request caps.                           |
| `mesh.resilience.retryBudget`      | Concurrent retries as a share of active requests.                              |
| `mesh.resilience.outlierDetection` | Ejection settings for Services with `replicas > 1`.                            |
| `mesh.resilience.grpcRetries`      | Timeouts, retry, and per-Service replay-safe methods for waypoint HTTPRoutes.  |
| `services.<name>.replicas`         | Deployment replicas (default 1).                                               |

## Verification

```bash
kubectl get ns ecommerce kafka --show-labels
kubectl get peerauthentication -A
kubectl get gateway,httproute,destinationrule -n ecommerce
kubectl get svc -n ecommerce -L istio.io/use-waypoint
kubectl get svc ecommerce-ingress-istio -n ecommerce   # expect ClusterIP
kubectl get tunnelbinding -n ecommerce
kubectl describe tunnelbinding ecommerce-ingress -n ecommerce   # DNS and config events

# Waypoint and ztunnel state:
istioctl ztunnel-config workloads -n ecommerce
istioctl proxy-config routes deploy/waypoint -n ecommerce -o json   # retryPolicy on listed methods only
istioctl proxy-config clusters deploy/waypoint -n ecommerce
istioctl proxy-config routes deploy/ecommerce-ingress-istio -n ecommerce -o json   # retryPolicy on GET only

curl -fsS -o /dev/null -w '%{http_code}\n' https://shop.phuchoang.sbs/
# Complete one checkout in the browser (login → cart → pay simulator → return).
```

Operators in non-enrolled platform namespaces (CNPG, MongoDB, Strimzi) and telemetry collectors reach enrolled pods without an Istio identity. Verify their reconciliation under `STRICT` during production acceptance and fix any broken management path explicitly rather than falling back to `PERMISSIVE`.

## Rollback

Disable marketplace `ingress.tunnel.enabled` (removes the public hostnames), `ingress.enabled`, or `mesh.waypoint.enabled` and sync. Cilium and Istio stay with talos-proxmox.
