## MODIFIED Requirements

### Requirement: Marketplace uses Istio Gateway API ingress

When ingress is enabled, the marketplace SHALL render a programmed `istio` Gateway and an HTTPRoute for web. Cloudflare Tunnel SHALL remain the public HTTPS front door; its in-cluster HTTP origin SHALL resolve the Istio-managed Gateway Service without a LAN VIP or an application TLS certificate. The dev host SHALL be `shop-dev` and the production host SHALL be `shop` through the existing overlay; the marketplace SHALL NOT route `pay` or `pay-dev`, which belong to the payment gateway.

#### Scenario: Web and simulator host routing

- **WHEN** a browser uses the configured shop hostname or a former simulator `pay` hostname
- **THEN** the Istio Gateway routes the shop request to the web Service, and the `pay` host and unknown hosts have no marketplace route

#### Scenario: Cloudflare origin is reconciled

- **WHEN** the tunnel binding and Gateway converge
- **THEN** the shop hostname resolves to the in-cluster Istio Gateway origin and forwarded HTTPS host/scheme semantics are preserved
