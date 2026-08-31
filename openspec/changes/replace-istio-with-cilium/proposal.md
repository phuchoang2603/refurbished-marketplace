## Why

Issue [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38): Istio ambient is four GitOps charts, a waypoint, Istio CNI, and L7 scrapes for an observe-only edge that this repo never used for mTLS, authz, or mesh tracing. The Talos cluster already runs Cilium with kube-proxy replacement, Gateway API, and Hubble. Replacing Istio with Cilium on both Talos and Colima leaves one dataplane and keeps Cloudflare Tunnel → Gateway API as the browser path.

## What Changes

- **BREAKING:** Remove Istio wrapper charts and Argo Applications (`base`, `istiod`, `cni`, `ztunnel`). Marketplace traffic no longer uses `gatewayClassName: istio` or `istio-waypoint`.
- Switch marketplace ingress to Kubernetes Gateway API with `gatewayClassName: cilium`. Keep the same host-based HTTPRoutes, `X-Forwarded-Proto` / `X-Forwarded-Host` filters, and Cloudflare Tunnel as the HTTPS front door.
- Provision the Cilium Gateway origin as ClusterIP (or equivalent in-cluster DNS) so `cloudflared` does not depend on L2 announcement VIPs. Repoint Cloudflare Public Hostnames to the new Service DNS.
- Delete ambient namespace labels, waypoint Gateway (`mesh.tpl`), Tilt label apply, and Istio L7 VMPodScrapes / Marketplace Istio RED dashboard.
- Treat Hubble L4 (Hubble UI / flows) plus existing app OTEL → VictoriaTraces as the observe path. No Cilium L7 visibility policies, no Hubble HTTP metrics requirement, no replacement Grafana RED dashboard in this change.
- Document expected Cilium Helm values for Talos (cluster-owned CNI, not Argo) and Colima/k3s (local bootstrap so Tilt/Argo match staging). Cilium stays the CNI: it is not an Argo app-of-apps child.
- Keep kafka in its own namespace with no L7 Cilium policies. Keep protocol-aware Service port names / `appProtocol` for Gateway and operators.
- Rewrite `docs/deployment/istio.md` into Cilium/Gateway docs; update GitOps and local-setup.

## Capabilities

### New Capabilities

- `cilium-ingress`: GitOps-managed Cilium Gateway API edge for marketplace browser traffic (and simulator), Cloudflare Tunnel HTTP origin, ClusterIP origin, TLS ownership, and ingress rollback.
- `cilium-observability`: Hubble L4 plus app OTEL as the post-Istio observe path; no waypoint, no Istio scrapes, no L7 mesh metrics requirement.

### Modified Capabilities

- `istio-ingress`: Retired. All Istio Gateway requirements are removed in favor of `cilium-ingress`.
- `istio-observability`: Retired. Ambient, waypoint, Istio chart pins, and Istio L7 telemetry requirements are removed. Protocol-aware Service ports move to `cilium-observability`.
- `argocd-gitops`: Drop GitOps-managed Istio Applications, istio-system PSS, and Istio sync waves. Ingress enablement targets Cilium Gateway resources. Kafka namespace separation remains (no L7 intercept of Strimzi TLS).
- `platform-observability`: Stop scraping Istio waypoint/ingress; drop Istio-as-RED-metrics. Traces remain app OTEL to VictoriaTraces.
- `distributed-tracing`: Keep “no mesh/Gateway proxy spans in the waterfall”; drop Istio-specific wording.

## Impact

- Touches `infra/charts/operators/istio/` (remove), `infra/argocd/app-of-apps/`, `infra/argocd/staging/root.yaml`, marketplace `ingress.tpl` / `mesh.tpl` / values, Tilt namespace labels, `infra/charts/cloudflare-tunnel/` comments, observability scrapes/dashboards/log exclude filters, and deployment docs.
- Adds documented Cilium values (Talos vs Colima); does not make Cilium an Argo Application.
- Does not change Go services, protobuf, databases, Kafka topics, or OTEL bootstrap.
- Follow-on mesh hardening and Argo canaries live in change `add-cilium-mesh-policy-and-canary` ([#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39)), not here.
