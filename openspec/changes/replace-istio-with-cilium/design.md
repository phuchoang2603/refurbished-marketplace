## Context

Local Colima and staging still install Istio ambient (`infra/charts/operators/istio/` pinned to 1.30.2) and enroll `ecommerce` via Argo `managedNamespaceMetadata` plus Tilt labels. Browser path is Cloudflare Tunnel → `Gateway` `gatewayClassName: istio` ClusterIP (`networking.istio.io/service-type: ClusterIP`) → HTTPRoutes for `shop*` / `pay*`. East–west L7 RED comes from `ecommerce-waypoint`. App traces already go OTEL → VictoriaTraces; mesh tracing is specified as absent.

The Talos cluster already runs Cilium (kube-proxy replacement, L2 announcements, `gatewayAPI.enabled`, Hubble). Cilium is the CNI: Argo cannot own it. Colima/k3s today is not Cilium.

Follow-on hardening is `add-cilium-mesh-policy-and-canary`.

## Goals / Non-Goals

**Goals:**

- One dataplane on Talos and Colima: Cilium CNI + Cilium Gateway API + Hubble L4.
- Same browser contract: Cloudflare HTTPS, HTTP origin, host-based routes, forwarded proto/host headers (payment callback 405 canary).
- Remove Istio from GitOps, charts, Tilt, observability scrapes/dashboards, and docs.
- Kafka namespace stays isolated from L7 proxy intercept.

**Non-Goals:**

- Strict mTLS, AuthorizationPolicy-equivalents, retries, circuit breakers, traffic splitting / canaries (next change).
- Hubble L7 HTTP/gRPC metrics, CiliumNetworkPolicy `http: [{}]` visibility, or a Grafana replacement for Marketplace Istio RED.
- Making Cilium an Argo Application.
- Changing Cloudflare Zero Trust Public Hostname UI into Git (only origin Service DNS changes).
- Production app-of-apps (still deferred).
- Mesh or Gateway proxy spans in VictoriaTraces.

## Decisions

### 1. Cilium is cluster-owned; this repo documents values, it does not sync the CNI

Keep Talos Cilium install outside Argo (bootstrap / Talos docs). Commit reference Helm values for Talos (the flags already in use: `kubeProxyReplacement`, `gatewayAPI`, Hubble, Talos cgroup/devices) and a separate Colima/k3s values file used by local bootstrap.

**Rationale:** CNI must exist before Argo. Chicken-egg if Cilium is a child Application.

**Alternatives considered:** Cilium Helm chart in app-of-apps (fails if the cluster cannot schedule pods without CNI); Istio+Cilium dual stack (rejected).

### 2. Colima must run Cilium too (option C)

Local k3s disables Traefik (already) and must run Cilium instead of default flannel so `GatewayClass/cilium` exists the same way as staging. Document Colima/k3s args + a one-shot or Tilt-owned Cilium Helm install with **non-Talos** values (`k8sServiceHost`/`Port`, cgroup, devices).

**Rationale:** otherwise local and staging diverge on the only remaining mesh/edge.

**Alternatives considered:** Istio on Colima, Cilium on Talos (faster, permanently forked DX).

### 3. Edge is still Gateway API in the marketplace chart

Keep `ingress.tpl` Gateway + two HTTPRoutes in `ecommerce`. Change `gatewayClassName` to `cilium`. Drop Istio ClusterIP annotation. Use `CiliumGatewayClassConfig` (parametersRef on GatewayClass, or a dedicated class) so the generated Service is ClusterIP for `cloudflared`. If ClusterIP is unavailable, document hitting the Service cluster DNS even if type is LoadBalancer — still must not require a LAN L2 VIP for the tunnel.

**Rationale:** routes are app-specific; Cloudflare only needs stable in-cluster DNS.

**Alternatives considered:** Cilium Ingress CRD (extra API); MetalLB/L2 VIP as the origin (breaks the tunnel-only design).

### 4. Delete waypoint and ambient enrollment

Remove `mesh.tpl`, `mesh.*` values, Argo/Tilt `istio.io/*` labels. No Cilium GAMMA mesh Gateway in this change.

**Rationale:** L7 waypoint existed for Istio RED; Hubble L4 + OTEL replace that observe story.

### 5. Observability: subtract Istio, do not add Hubble Prometheus as a requirement

Remove `istioScrapes`, Istio RED dashboard, `istio-proxy` log excludes if unused. Hubble UI remains the L4 flow UI (already enabled on Talos; enable equivalently on Colima). App traces unchanged.

**Rationale:** user chose not to recreate L7 RED.

### 6. Protocol-aware Service ports stay

Keep Helm `services.*.protocol` port names / appProtocol. Useful for Cilium Gateway and future L7 policy. Not used to drive Hubble HTTP metrics in this change.

### 7. Cut over origin DNS in Cloudflare last

Service name will change off `ecommerce-ingress-istio`. Sequence: Cilium Gateway Accepted → new Service Ready → update Public Hostnames → delete Istio. Reverse for rollback if shop/pay break (especially hosted-payment POST → 405).

## Risks / Trade-offs

| Risk                                                                               | Mitigation                                                                |
| ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| Colima Cilium bootstrap is the hard local DX piece                                 | Spike k3s + Cilium + GatewayClass before deleting Istio locally           |
| `CiliumGatewayClassConfig` ClusterIP not supported on the installed Cilium version | Fall back to in-cluster Service DNS; never require L2 VIP for cloudflared |
| Header filters unimplemented on Cilium Gateway                                     | Spike `RequestHeaderModifier`; payment callback is the canary             |
| Gateway API CRD channel mismatch                                                   | Pin CRDs to what Cilium documents for that version                        |
| Dual GatewayClasses during migrate                                                 | Only one parent for shop/pay hosts at a time                              |
| Hubble UI not GitOps-exposed                                                       | Document port-forward / in-cluster access; not a Grafana requirement      |

## Migration Plan

1. Confirm Talos `GatewayClass/cilium` Accepted; Hubble relay/UI up.
2. Add Colima Cilium bootstrap + values; verify `GatewayClass/cilium` locally.
3. Add ClusterIP GatewayClass config; switch marketplace `ingress.tpl`; keep Istio installed until the new Gateway is Accepted.
4. Point Cloudflare Public Hostnames at the Cilium Gateway Service; verify shop + pay callback.
5. Remove waypoint, ambient labels, Istio Argo apps/charts, Istio scrapes/dashboard.
6. Rewrite docs; OpenSpec validate.

Rollback: restore Istio Gateway class and Cloudflare origin DNS; re-enable Istio apps before deleting Cilium Gateway routes.
