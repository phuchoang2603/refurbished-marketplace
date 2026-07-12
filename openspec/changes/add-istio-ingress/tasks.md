## 1. Chart ingress surface

- [x] 1.1 Add marketplace chart values for `ingress` (enabled flag, Gateway name, web hostname, simulator hostname, listener port).
- [x] 1.2 Add Helm templates for Istio `Gateway` (`gatewayClassName: istio`) distinct from the existing waypoint Gateway.
- [x] 1.3 Add `HTTPRoute` for `web` attached to the ingress Gateway.
- [x] 1.4 Add `HTTPRoute` for `payment-gateway-simulator` on a distinct hostname attached to the same Gateway.
- [x] 1.5 Keep chart defaults with `ingress.enabled: false` so Tilt remains unchanged.

## 2. Staging GitOps enablement

- [x] 2.1 Enable ingress on the staging marketplace Argo Application / values overlay with concrete Cloudflare DNS hostnames.
- [x] 2.2 Set staging `HOSTED_PAYMENT_BASE_URL` to the Cloudflare-facing simulator HTTPS URL.
- [x] 2.3 Confirm production overlays leave ingress disabled.

## 3. Docs and TLS ownership

- [x] 3.1 Update `docs/deployment/istio.md` with edge Gateway vs waypoint, Cloudflare Tunnel origin wiring, and verification commands.
- [x] 3.2 Document TLS at Cloudflare + HTTP origin at Istio, and rollback steps to disable ingress.

## 4. Staging verification

- [ ] 4.1 Sync staging and confirm Gateway/HTTPRoute resources become Accepted/healthy and a LoadBalancer or Service address is available for the tunnel origin.
- [ ] 4.2 Open the staging web URL through Cloudflare → Istio and exercise product → cart → checkout → payment (including simulator redirect).
- [ ] 4.3 Confirm requests enter via the Istio gateway path (Service/metrics/logs) and unmatched hosts behave as documented.
- [ ] 4.4 Disable ingress via GitOps and confirm marketplace Gateway/HTTPRoute resources are removed; Tilt port-forwards still work locally.

## 5. Close-out

- [x] 5.1 Run OpenSpec validation for `add-istio-ingress`.
- [x] 5.2 Link or update GitHub issue #19 acceptance criteria against the implemented behavior.
