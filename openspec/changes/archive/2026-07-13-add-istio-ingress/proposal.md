## Why

Issue [#19](https://github.com/phuchoang2603/refurbished-marketplace/issues/19): staging now has an observe-only ambient mesh, but browser traffic still has no GitOps-managed edge path into `web`. Without an Istio-managed ingress, end-to-end product → cart → checkout → payment verification stays blocked, and the hosted payment simulator cannot be reached by browsers through a documented public URL.

## What Changes

- Add GitOps-managed Istio edge routing for staging using the Kubernetes Gateway API (`Gateway` + `HTTPRoute`) with Istio as the implementation (`gatewayClassName: istio`).
- Route external HTTP traffic to the marketplace `web` Service.
- Explicitly expose the hosted `payment-gateway-simulator` on a browser-reachable edge route (or document a deliberate deferral with reason — default plan is to expose it).
- Configure staging `HOSTED_PAYMENT_BASE_URL` to the Cloudflare-facing simulator URL so hosted-payment redirects work from a real browser.
- Document TLS ownership (Cloudflare terminates HTTPS; Istio Gateway is HTTP origin via in-cluster Cloudflare Tunnel), migration, and rollback.
- Deploy GitOps-managed in-cluster `cloudflared` (token from Doppler/ESO); Public Hostnames stay in the Cloudflare dashboard.
- Keep local Tilt port-forward behavior unchanged; do not require a local mesh ingress mode.
- Leave production ingress disabled until staging edge routing is verified.

## Capabilities

### New Capabilities

- `istio-ingress`: GitOps-managed Istio edge gateway and HTTP routes for marketplace browser traffic (and simulator), in-cluster Cloudflare Tunnel connector, TLS ownership, and rollback.

### Modified Capabilities

- `argocd-gitops`: Staging GitOps overlays enable Istio ingress resources, browser-reachable simulator URL configuration, and the Cloudflare Tunnel Application.
- `istio-observability`: Ingress is no longer permanently deferred; observe-only mesh remains non-disruptive, while edge routing becomes a separate, explicit enablement path.

## Impact

- Touches `infra/charts/refurbished-marketplace/` (ingress templates/values), `infra/charts/cloudflare-tunnel/`, `infra/argocd/staging/` (enablement overlays), and docs for edge/TLS/secrets/GitOps.
- Uses existing Gateway API CRDs and `GatewayClass/istio` already Accepted on staging.
- Adds in-cluster `cloudflared` with tunnel token from Doppler (`CLOUDFLARE_TUNNEL_TOKEN`).
- Does not change Go service business logic, protobuf contracts, or Tilt local workflows.
- Unblocks the deferred browser flow verification called out in the observe-baseline docs.
- Cloudflare Zero Trust Public Hostname UI mapping remains outside Git (points at Gateway ClusterIP Service DNS).
