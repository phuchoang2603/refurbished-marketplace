## Purpose

Define Cilium Gateway API edge ingress for marketplace browser traffic, with Cloudflare Tunnel as the public HTTPS front door for local Colima (`.dev` hosts) and staging (production hostnames).

## ADDED Requirements

### Requirement: GitOps-managed Cilium edge gateway

The system SHALL provide GitOps-managed Kubernetes Gateway API resources that use Cilium (`gatewayClassName: cilium`, or a documented Cilium GatewayClass with ClusterIP parameters) as the edge implementation for marketplace browser traffic when ingress is enabled (chart defaults for local; staging overlays for production hostnames).

#### Scenario: Sync creates edge Gateway

- **WHEN** marketplace values enable ingress and the chart is applied (Tilt locally or Argo CD on staging)
- **THEN** a `Gateway` with a Cilium GatewayClass exists and is Accepted for HTTP browser entry

#### Scenario: No Istio waypoint gateway

- **WHEN** marketplace ingress is enabled after Istio removal
- **THEN** the chart does not render an `istio-waypoint` Gateway or HBONE listener

#### Scenario: Chart defaults keep ingress on for local .dev hosts

- **WHEN** the marketplace chart renders with default values
- **THEN** ingress Gateway and HTTPRoute resources are rendered for `shop-dev.phuchoang.sbs` / `pay-dev.phuchoang.sbs`

### Requirement: Browser traffic reaches web through Cilium Gateway

The system SHALL route external HTTP traffic that matches the configured web hostname to the marketplace `web` Service through the Cilium-managed edge Gateway.

#### Scenario: Web host routes to web service

- **WHEN** a browser request hits the Cilium edge Gateway with the configured web hostname
- **THEN** the request is routed to the `web` Service in the marketplace namespace

#### Scenario: Unmatched host or path follows documented behavior

- **WHEN** a request reaches the edge Gateway with a host or path that is not configured
- **THEN** the request is rejected or not routed to marketplace backends according to the documented Gateway/HTTPRoute rules

### Requirement: Hosted payment simulator edge exposure

The system SHALL expose the hosted `payment-gateway-simulator` on a Cilium-managed browser-reachable route via a distinct hostname on the same ingress Gateway.

#### Scenario: Simulator is reachable through Cilium Gateway

- **WHEN** ingress is enabled with simulator routing configured
- **THEN** an HTTPRoute sends matching browser traffic for the simulator hostname to the `payment-gateway-simulator` Service

#### Scenario: Web uses browser-reachable simulator base URL

- **WHEN** staging ingress overlays are applied
- **THEN** `HOSTED_PAYMENT_BASE_URL` targets the Cloudflare-facing simulator HTTPS URL rather than cluster-only DNS or localhost

#### Scenario: Origin scheme headers preserve HTTPS callbacks

- **WHEN** HTTPRoutes for web and simulator are rendered
- **THEN** they set `X-Forwarded-Proto: https` and `X-Forwarded-Host` to the route hostname so hosted-payment callbacks are not rewritten to HTTP

### Requirement: Cloudflare Tunnel is the public front door

Local and staging marketplace edges SHALL assume Cloudflare Tunnel as the public HTTPS front door and the Cilium Gateway as the HTTP origin. The repository SHALL deploy an in-cluster `cloudflared` connector through Argo CD and SHALL NOT require a marketplace TLS certificate on the Cilium Gateway for this path. The Gateway Service SHALL be reachable in-cluster as ClusterIP (or documented equivalent Service DNS) without requiring an L2 announcement VIP for the tunnel.

#### Scenario: Origin is HTTP behind Cloudflare

- **WHEN** ingress is enabled for Cloudflare Tunnel access
- **THEN** the Cilium Gateway listens for HTTP from the in-cluster tunnel connector and does not require a marketplace TLS Secret for browser access

#### Scenario: Public hostnames match route hostnames

- **WHEN** Cloudflare Public Hostnames are configured for web and simulator
- **THEN** those hostnames match the Gateway/HTTPRoute hostname values and the origin URL uses the Cilium Gateway Service DNS (not `ecommerce-ingress-istio`)

#### Scenario: cloudflared is GitOps-managed

- **WHEN** the local or staging root Application syncs from Git
- **THEN** Argo CD manages a `cloudflare-tunnel` Application that runs `cloudflared` with a tunnel token sourced from External Secrets

### Requirement: TLS termination ownership is documented

The repository SHALL document that marketplace browser TLS terminates at Cloudflare, with HTTP from Cloudflare Tunnel to the Cilium Gateway origin.

#### Scenario: Contributor finds TLS ownership

- **WHEN** a contributor reads marketplace ingress deployment docs after this change
- **THEN** the docs state that local and staging terminate TLS at Cloudflare and use HTTP between the tunnel and Cilium Gateway

### Requirement: Ingress rollback is documented

The system SHALL document rollback steps that disable Cilium marketplace ingress without requiring application code changes.

#### Scenario: Ingress disabled

- **WHEN** ingress enablement is turned off and synced
- **THEN** marketplace browser traffic no longer depends on the Cilium edge Gateway for that environment

### Requirement: Cilium CNI is cluster-owned

The repository SHALL document expected Cilium Helm values for Talos and Colima and SHALL NOT manage the Cilium agent as an Argo CD Application. Local Colima SHALL run Cilium so `GatewayClass` for Cilium is available the same way as staging.

#### Scenario: Talos Cilium remains bootstrap

- **WHEN** staging GitOps syncs after this change
- **THEN** no Argo Application installs or upgrades the cluster CNI (Cilium)

#### Scenario: Colima has Cilium GatewayClass

- **WHEN** local Kubernetes is started per documented local-setup
- **THEN** a Cilium GatewayClass is Accepted before marketplace ingress is applied
