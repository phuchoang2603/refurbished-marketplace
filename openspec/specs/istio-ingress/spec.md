# istio-ingress Specification

## Purpose

TBD - created by archiving change add-istio-ingress. Update Purpose after archive.

## Requirements

### Requirement: GitOps-managed Istio edge gateway

The system SHALL provide GitOps-managed Kubernetes Gateway API resources that use Istio (`gatewayClassName: istio`) as the edge implementation for marketplace browser traffic in staging when ingress is enabled.

#### Scenario: Staging sync creates edge Gateway

- **WHEN** staging marketplace values enable ingress and ArgoCD syncs
- **THEN** a `Gateway` with `gatewayClassName: istio` exists and is Accepted for HTTP browser entry

#### Scenario: Edge Gateway is distinct from waypoint

- **WHEN** marketplace ambient waypoint and ingress are both enabled
- **THEN** the ingress `Gateway` uses `gatewayClassName: istio` and the waypoint `Gateway` continues to use `gatewayClassName: istio-waypoint`

#### Scenario: Chart defaults keep ingress off for Tilt

- **WHEN** the marketplace chart renders with default values
- **THEN** ingress Gateway and HTTPRoute resources are not rendered

### Requirement: Browser traffic reaches web through Istio

The system SHALL route external HTTP traffic that matches the configured web hostname to the marketplace `web` Service through the Istio-managed edge Gateway.

#### Scenario: Web host routes to web service

- **WHEN** a browser request hits the Istio edge Gateway with the configured web hostname
- **THEN** the request is routed to the `web` Service in the marketplace namespace

#### Scenario: Unmatched host or path follows documented behavior

- **WHEN** a request reaches the edge Gateway with a host or path that is not configured
- **THEN** the request is rejected or not routed to marketplace backends according to the documented Gateway/HTTPRoute rules

### Requirement: Hosted payment simulator edge exposure

The system SHALL expose the hosted `payment-gateway-simulator` on an Istio-managed browser-reachable route in staging via a distinct hostname on the same ingress Gateway.

#### Scenario: Simulator is reachable through Istio

- **WHEN** staging ingress is enabled with simulator routing configured
- **THEN** an HTTPRoute sends matching browser traffic for the simulator hostname to the `payment-gateway-simulator` Service

#### Scenario: Web uses browser-reachable simulator base URL

- **WHEN** staging ingress overlays are applied
- **THEN** `HOSTED_PAYMENT_BASE_URL` targets the Cloudflare-facing simulator HTTPS URL rather than cluster-only DNS or Tilt localhost

### Requirement: Cloudflare Tunnel is the public front door

The staging marketplace edge SHALL assume Cloudflare Tunnel as the public HTTPS front door and the Istio Gateway as the HTTP origin. The repository SHALL deploy an in-cluster `cloudflared` connector through Argo CD and SHALL NOT require a marketplace TLS certificate on the Istio Gateway for this path.

#### Scenario: Origin is HTTP behind Cloudflare

- **WHEN** staging ingress is enabled for Cloudflare Tunnel access
- **THEN** the Istio Gateway listens for HTTP from the in-cluster tunnel connector and does not require a marketplace TLS Secret for browser access

#### Scenario: Public hostnames match route hostnames

- **WHEN** Cloudflare Public Hostnames are configured for web and simulator
- **THEN** those hostnames match the Gateway/HTTPRoute hostname values used by Istio

#### Scenario: cloudflared is GitOps-managed

- **WHEN** the staging root Application syncs from Git
- **THEN** Argo CD manages a `cloudflare-tunnel` Application that runs `cloudflared` with a tunnel token sourced from External Secrets

### Requirement: TLS termination ownership is documented

The repository SHALL document that marketplace browser TLS terminates at Cloudflare, with HTTP from Cloudflare Tunnel to the Istio Gateway origin.

#### Scenario: Contributor finds TLS ownership

- **WHEN** a contributor reads Istio deployment docs after this change
- **THEN** the docs state that staging terminates TLS at Cloudflare and uses HTTP between the tunnel and Istio

### Requirement: Ingress rollback is documented

The system SHALL document rollback steps that disable Istio marketplace ingress without requiring application code changes. Local Tilt port-forwards remain the local access path when ingress is disabled.

#### Scenario: Ingress disabled

- **WHEN** staging ingress enablement is turned off and synced
- **THEN** marketplace browser traffic no longer depends on the Istio edge Gateway and operators can use the documented local or prior access path
