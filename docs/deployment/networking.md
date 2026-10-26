# Networking, Istio ambient, and ingress

The **talos-proxmox** platform owns the cluster network. Cilium is the CNI and L2/LB IPAM provider; its Gateway controller, SPIRE authentication, and L7 proxy are disabled. Istio (base, istiod, CNI, ztunnel) supplies ambient mesh transport and the `istio` / `istio-waypoint` GatewayClasses. This repo installs neither and must not duplicate their Helm values.

This repo consumes that dataplane:

- The marketplace root renders the `ecommerce` and `kafka` Namespaces with `istio.io/dataplane-mode: ambient`. Child Applications never create or relabel them, and no platform namespace is enrolled.
- Each namespace has a `STRICT` `PeerAuthentication` (marketplace chart for `ecommerce`, Kafka chart for `kafka`). Kafka keeps its own broker TLS; ambient mTLS is transport protection, not a Kafka replacement.
- The marketplace chart renders the browser `Gateway`/`HTTPRoute`s (`gatewayClassName: istio`), a service waypoint, and per-service `DestinationRule`s.

No CiliumNetworkPolicy is rendered. In prod, `NetworkPolicy/web-payment-gateway` limits web ingress to local namespace pods and the gateway notifier (`app.kubernetes.io/name: notifier`, `payment-gateway/release: notifier`) in namespace `payment-gateway`, on TCP 8080 and HBONE 15008, plus ambient probe traffic from `169.254.7.127`. The local namespace rule preserves ingress Gateway/browser and service access; egress is unchanged. Application traces and RED metrics are OTLP/HTTP exports to VMAgent and VTAgent in `observability`; see [observability.md](observability.md).

## Edge

Browser traffic: Cloudflare Tunnel → Istio Gateway → `web`.

| Env  | Hostname             | Backend |
| ---- | -------------------- | ------- |
| prod | `shop.phuchoang.sbs` | `web`   |

Istio provisions `Service/ecommerce-ingress-istio` for `Gateway/ecommerce-ingress`. The Gateway's `networking.istio.io/service-type: ClusterIP` annotation keeps it off the Cilium L2 pool; the tunnel reaches it through in-cluster DNS:

`http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`

HTTPRoutes set `X-Forwarded-Proto: https` and `X-Forwarded-Host` so browser redirects preserve the external HTTPS scheme and shop hostname.

The chart renders `TunnelBinding/ecommerce-ingress` (`ingress.tunnel`) with the shop hostname and an explicit `target` on that origin, because the generated Service also exposes the Istio status port. The `talos-proxmox` Cloudflare operator adds it to the environment's `ClusterTunnel/talos-proxmox`, restarts `cloudflared`, and creates each hostname's proxied CNAME and `_managed.<hostname>` ownership TXT record. Deleting the binding, or disabling `ingress.tunnel`, removes the records. Nothing is configured in the Cloudflare dashboard.

The operator refuses a hostname that already has a DNS record it does not own. Delete such a record in Cloudflare and the binding converges on its next retry.

TLS terminates at Cloudflare. No marketplace TLS Secret on the Gateway. Do not reuse the platform Argo CD or HyperDX Gateways for shop.

The gateway notifier calls only `http://web.ecommerce.svc.cluster.local:8080/webhooks/payment-gateway`. An Istio `DENY` AuthorizationPolicy attached to `Gateway/ecommerce-ingress` blocks that path and subpaths for every method before the public shop catch-all can forward them. The internal Service call bypasses this ingress Gateway. Web forwards the unmodified body and signature to Checkout; Checkout verifies the HMAC. The notifier and Edge are ambient-enrolled in prod, so mesh `STRICT` remains enabled.

Production cluster validation of browser access, notifier delivery, and public webhook rejection is deferred with the [rollout gate](../order-placement.md#rollout-gate).

## Waypoint and resilience

`Gateway/waypoint` (`istio-waypoint`, `istio.io/waypoint-for: service`) handles L7 traffic for the gRPC Services listed in `mesh.resilience.services`; those Services carry `istio.io/use-waypoint: waypoint`. `web`, Kafka, MongoDB, Meilisearch, and CNPG stay on ztunnel only. The ingress Gateway is itself an Envoy that enforces the shop route policies, and its web backend has no waypoint, so no Service sets `istio.io/ingress-use-waypoint`. Add that label only if a waypoint-enrolled Service becomes an ingress backend.

Each listed Service gets a `DestinationRule` with connection and pending-request caps and a retry budget. Outlier ejection is rendered only for Services with `replicas > 1`, so a sole endpoint is never ejected.

Retry rules depend on two platform prerequisites from `talos-proxmox`: mesh-wide default HTTP retries disabled (`meshConfig.defaultHttpRetryPolicy.attempts: 0`) and the Gateway API Experimental CRDs, which admit `HTTPRoute.rules[].retry`.

- **Ingress reads** (`ingress.reads`): shop `GET`s retry twice on connect failures, refused streams, and 502/503, with a 9s per-attempt and 30s total budget.
- **Ingress writes** (`ingress.writes`): every other request, including checkout and cart POSTs, has a 30s total / 25s backend budget and never retries.
- **Waypoint gRPC** (`mesh.resilience.grpcRetries`): each listed Service gets a `<name>-mesh` HTTPRoute attached at the waypoint. Exact unary method paths for lookups and `SubmitCheckout` retry twice with a 3s per-attempt and 10s total budget; an identical `SubmitCheckout` (same buyer, intent key, and payload) returns the original checkout. Istio retries connect failures, refused streams, and gRPC `UNAVAILABLE`/`CANCELLED` responses; it does not retry other gRPC statuses. The catch-all rule routes every other method, including cart mutations and `HandleGatewayWebhook`, without retries or a timeout.

Proxy retries do not replace the checkout saga's inbox/outbox replay, deadlines, or compensation, and Istio never retries Kafka delivery.

| Value                              | Effect                                                                         |
| ---------------------------------- | ------------------------------------------------------------------------------ |
| `ingress.reads`                    | Method, timeouts, and retry for the shop read rule.                            |
| `ingress.writes`                   | Timeouts for the unretried shop catch-all rule.                                |
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
# Complete one checkout in the browser (login → cart → hosted gateway → return (prod only)).
```

Under `STRICT`, ztunnel rejects plaintext from non-enrolled pods. `talos-proxmox` therefore enrolls the CNPG and Strimzi operators and the `otel-cluster` scraper in ambient at the pod level, so their calls to enrolled pods use mesh mTLS; the MongoDB operator needs no pod access. Mesh traffic reaches enrolled pods on HBONE port 15008, which Strimzi's generated NetworkPolicies do not allow, so the Kafka chart adds `NetworkPolicy/allow-ambient` for Strimzi pods, which also admits kubelet probes from ambient's link-local source `169.254.7.127`. Fix any other broken management path the same way rather than falling back to `PERMISSIVE`.

## Rollback

Disable marketplace `ingress.tunnel.enabled` (removes the public hostnames), `ingress.enabled`, or `mesh.waypoint.enabled` and sync. Cilium and Istio stay with talos-proxmox.
