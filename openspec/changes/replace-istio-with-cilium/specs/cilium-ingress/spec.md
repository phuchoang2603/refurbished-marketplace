## Purpose

Define Cilium Gateway API edge ingress for marketplace browser traffic, with Cloudflare Tunnel as the public HTTPS front door on Talos using the same shop/pay hostnames as production.

## ADDED Requirements

### Requirement: GitOps-managed Cilium edge gateway

The system SHALL provide GitOps-managed Kubernetes Gateway API resources that use Cilium (`gatewayClassName: cilium`, or a documented Cilium GatewayClass with ClusterIP parameters) as the edge implementation for marketplace browser traffic when ingress is enabled. Chart defaults SHALL use the cluster shop/pay hostnames (not a Colima `.dev` overlay).

#### Scenario: Sync creates edge Gateway

- **WHEN** marketplace values enable ingress and Argo CD syncs the marketplace chart on Talos
- **THEN** a `Gateway` with a Cilium GatewayClass exists and is Accepted for HTTP browser entry

#### Scenario: No Istio waypoint gateway

- **WHEN** marketplace ingress is enabled after Istio removal
- **THEN** the chart does not render an `istio-waypoint` Gateway or HBONE listener

#### Scenario: Chart defaults keep ingress on for cluster hosts

- **WHEN** the marketplace chart renders with default values
- **THEN** ingress Gateway and HTTPRoute resources are rendered for the cluster shop/pay hostnames (not a Colima-only `.dev` pair)

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

- **WHEN** ingress is enabled with simulator routing configured
- **THEN** `HOSTED_PAYMENT_BASE_URL` targets the Cloudflare-facing simulator HTTPS URL rather than cluster-only DNS or localhost

#### Scenario: Origin scheme headers preserve HTTPS callbacks

- **WHEN** HTTPRoutes for web and simulator are rendered
- **THEN** they set `X-Forwarded-Proto: https` and `X-Forwarded-Host` to the route hostname so hosted-payment callbacks are not rewritten to HTTP

### Requirement: Cloudflare Tunnel is the public front door

Talos marketplace edges SHALL assume Cloudflare Tunnel as the public HTTPS front door and the Cilium Gateway as the HTTP origin. The repository SHALL deploy an in-cluster `cloudflared` connector through Argo CD and SHALL NOT require a marketplace TLS certificate on the Cilium Gateway for this path. The Gateway Service SHALL be reachable in-cluster as ClusterIP (or documented equivalent Service DNS) without requiring an L2 announcement VIP for the tunnel.

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

The repository SHALL document expected Cilium Helm values for Talos and SHALL NOT manage the Cilium agent as an Argo CD Application. Marketplace ingress SHALL be applied by Argo CD on that cluster so `GatewayClass` for Cilium is available without Colima/k3s or Tilt.

#### Scenario: Talos Cilium remains bootstrap

- **WHEN** GitOps syncs after this change
- **THEN** no Argo Application installs or upgrades the cluster CNI (Cilium)

#### Scenario: Argo uses Talos GatewayClass

- **WHEN** the Talos marketplace Application syncs with ingress enabled
- **THEN** a Cilium GatewayClass is Accepted and the marketplace Gateway is bound to it

### Requirement: Tilt is not the cluster deploy path

Documented developer workflow SHALL deploy marketplace workloads with Argo CD and GHCR, not Tilt `helm`/`docker_build`. templ and Tailwind MAY still be generated on the laptop and committed; they SHALL NOT require a Tilt watch attached to cluster pods.

#### Scenario: No Tilt apply for marketplace

- **WHEN** a contributor follows local-setup after this change
- **THEN** they are not required to run `tilt up` to install Argo, build images, or apply the marketplace chart

#### Scenario: Branch tracking uses git and GHCR

- **WHEN** the Talos root Application `targetRevision` is a git branch whose images were published by CI
- **THEN** marketplace pods pull `global.imageTag` matching that commit SHA from GHCR
