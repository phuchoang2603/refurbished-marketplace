## Context

Staging already runs Istio ambient (`base` / `istiod` / `cni` / `ztunnel`) with marketplace waypoint enrollment via `mesh.tpl`. Gateway API CRDs are present; `GatewayClass` objects `istio`, `istio-remote`, and `istio-waypoint` are Accepted.

Browser edge today:

- Marketplace `web` and `payment-gateway-simulator` have no GitOps-managed external edge route.
- Local Tilt uses port-forwards (`8080` → web, `8097` → simulator).
- Staging Argo overlay sets `HOSTED_PAYMENT_BASE_URL=http://payment-gateway-simulator:8097`, which is cluster-DNS-only and does not work for browser redirects.

Issue #19 asks to move marketplace edge routing onto Istio after the observe baseline. The intended public front door is Cloudflare Tunnel pointing at the Istio Gateway origin.

## Goals / Non-Goals

**Goals:**

- GitOps-managed Istio edge for staging that sends browser HTTP to `web`.
- Explicit browser path to `payment-gateway-simulator` and a matching public `HOSTED_PAYMENT_BASE_URL`.
- GitOps-managed in-cluster `cloudflared` connector (token via Doppler/ESO).
- Documented TLS ownership (Cloudflare) and ingress rollback.
- Preserve Tilt port-forwards with ingress disabled by default in chart values.

**Non-Goals:**

- Production ingress enablement in this change.
- Strict mTLS, AuthorizationPolicy, retries, or canaries at the edge.
- Local ambient/ingress mode for Tilt.
- Marketplace origin TLS / cert-manager (Cloudflare terminates TLS in front of HTTP Istio).
- Managing Cloudflare Zero Trust Public Hostname UI config in Git (only the `cloudflared` connector is GitOps-managed; host→origin mapping stays in the Cloudflare dashboard unless added later).
- Proxmox LXC `cloudflared` as the marketplace edge connector.

## Decisions

### 1. Use Kubernetes Gateway API with `gatewayClassName: istio`

Prefer `Gateway` + `HTTPRoute` over Istio `Gateway`/`VirtualService` CRDs.

**Rationale:** staging already uses Gateway API for the ambient waypoint (`istio-waypoint`). Istio documents Gateway API as the preferred ingress path. One API family for waypoint + edge reduces cognitive load.

**Alternatives considered:** classic Istio `Gateway`/`VirtualService` (works, but duplicates models); generic Kubernetes `Ingress` (not the mesh edge model this change targets).

### 2. Own edge resources in the marketplace chart behind `ingress.enabled`

Render from the marketplace Helm chart (extend `mesh.tpl` or add `ingress.tpl`):

- One `Gateway` (class `istio`) with HTTP listener(s).
- `HTTPRoute` → `web` Service.
- `HTTPRoute` → `payment-gateway-simulator` Service.
- Values for hostnames, ports, and enable flags.

Staging Argo Application values set `ingress.enabled: true` and host/URL overlays. Chart defaults keep `ingress.enabled: false` for Tilt.

**Rationale:** routes are app-specific; the chart already owns ambient labels and the waypoint Gateway. A separate platform chart would still need marketplace host/service knowledge.

**Alternatives considered:** dedicated `istio-ingress` platform chart + app-only HTTPRoutes (cleaner long-term split, more Argo apps for this repo size).

### 3. Dedicated edge Gateway, not the waypoint Gateway

Keep `ecommerce-waypoint` (`gatewayClassName: istio-waypoint`, HBONE/15008) for east-west L7. Add a separate ingress `Gateway` (`gatewayClassName: istio`) that provisions a north-south proxy and LoadBalancer Service.

**Rationale:** waypoint and ingress GatewayClasses have different controllers and listener semantics; mixing them breaks ambient waypoint behavior.

### 4. In-cluster Cloudflare Tunnel (cloudflared) via Argo CD; HTTP-only at Istio

Staging browser TLS terminates at **Cloudflare**. `cloudflared` runs **in-cluster** (Argo Application `staging-cloudflare-tunnel`, namespace `cloudflare-tunnel`) using a remotely managed tunnel token from Doppler/ESO. Public Hostnames in Cloudflare Zero Trust point at the Istio Gateway **ClusterIP** Service. No cert Secret / cert-manager on the marketplace Gateway.

| Layer                                 | Staging v1                                         |
| ------------------------------------- | -------------------------------------------------- |
| Browser → Cloudflare                  | HTTPS (Cloudflare-managed cert)                    |
| Cloudflare → cloudflared (in-cluster) | Cloudflare edge to tunnel connector                |
| cloudflared → Istio Gateway           | HTTP to `ecommerce-ingress-istio.ecommerce.svc:80` |
| Gateway → `web` / simulator           | cluster HTTP                                       |

**Rationale:** GitOps-managed connector, stable Service DNS origin, no MetalLB dependency for public access. Proxmox LXC tunnel is no longer the target path.

**Alternatives considered:** Proxmox LXC `cloudflared` (kept out of GitOps); TLS at Istio with a Secret (unnecessary with tunnel).

### 5. Simulator on a distinct hostname

Second hostname on the same ingress Gateway → `payment-gateway-simulator:8097`. Staging `HOSTED_PAYMENT_BASE_URL` becomes `https://pay.<domain>` (browser URL via Cloudflare), not cluster DNS.

**Rationale:** Cloudflare Tunnel maps cleanly to one Public Hostname per service; absolute redirect URLs stay simple; no path-rewrite risk for the simulator app.

**Alternatives considered:** path-prefix under the web host (possible, but messier redirects and CF config); defer simulator edge (blocks real checkout verification).

### 6. Cloudflare + Istio is the marketplace edge path

Marketplace browser traffic enters only via Cloudflare Tunnel → Istio Gateway. No alternate project ingress path is in scope. Document verification via Gateway Service / metrics and tunnel host wiring.

### 7. Real DNS hostnames via Cloudflare (values-driven)

Staging values supply real hostnames managed in Cloudflare DNS/tunnel (e.g. `shop.<domain>`, `pay.<domain>`). Chart stays values-driven — no hard-coded domain. Skip sslip.io and Host-header-only as the primary scheme.

### 8. Ingress `Gateway` lives in `ecommerce` with the HTTPRoutes

Keep Gateway + HTTPRoutes in `ecommerce` (marketplace chart / Argo destination).

**Rationale:** staging marketplace Application targets `ecommerce`; cross-namespace Gateway resources would need a second Argo app or relaxed destination rules. One chart/namespace is enough for this change. Cloudflare Tunnel only needs a stable origin address (the Gateway Service), not a dedicated namespace.

**Alternatives considered:** dedicated `istio-ingress` namespace + `parentRefs` from ecommerce HTTPRoutes (cleaner platform split later, extra Argo app now).

## Risks / Trade-offs

| Risk                                              | Mitigation                                                                                   |
| ------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Tunnel token missing in Doppler                   | Document `CLOUDFLARE_TUNNEL_TOKEN` bootstrap; ExternalSecret blocks Deployment until present |
| Tunnel Host header must match HTTPRoute hostnames | Configure CF Public Hostnames to the same names as Gateway/HTTPRoute `hostnames`             |
| Ambient labels missing after NS recreate          | Keep `mesh.ambient` / waypoint values; verify namespace labels after sync                    |
| Simulator and web on split hosts                  | Related DNS names; web already owns redirect/callback URL construction                       |
| Gateway status drift in Argo                      | IgnoreDifferences for known status fields if needed                                          |
| Operators confuse waypoint vs ingress Gateway     | Docs table: class, port, purpose                                                             |

## Migration Plan

1. Confirm `GatewayClass/istio` Accepted and marketplace Services healthy.
2. Add chart templates/values; keep default disabled.
3. Enable on staging via Argo values; sync Gateway as ClusterIP Service.
4. Add `CLOUDFLARE_TUNNEL_TOKEN` to Doppler `prd`; sync `staging-cloudflare-tunnel`.
5. In Cloudflare Zero Trust, set Public Hostnames `shop.phuchoang.sbs` and `pay.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`.
6. Set `HOSTED_PAYMENT_BASE_URL` to `https://pay.phuchoang.sbs`; exercise checkout payment redirect.
7. Rollback: disable marketplace ingress and/or remove the cloudflare-tunnel Application; Tilt port-forward still works locally.

## Resolved questions (Cloudflare Tunnel)

1. **Hostname scheme:** real DNS in Cloudflare (not sslip.io / Host-header-only).
2. **TLS:** HTTP-only at Istio; TLS at Cloudflare — no marketplace cert required.
3. **Simulator:** separate hostname on the same Gateway.
4. **Gateway namespace:** `ecommerce` with the HTTPRoutes (dedicated `istio-ingress` deferred).
5. **Tunnel connector:** in-cluster `cloudflared` via Argo CD (not Proxmox LXC).
