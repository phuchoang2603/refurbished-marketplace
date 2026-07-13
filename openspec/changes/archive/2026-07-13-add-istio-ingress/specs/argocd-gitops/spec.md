## ADDED Requirements

### Requirement: Staging Istio ingress enablement

The staging ArgoCD marketplace Application SHALL be able to enable Istio edge Gateway API resources through Helm value overlays without requiring Tilt.

#### Scenario: Staging overlay enables ingress

- **WHEN** staging marketplace values set ingress enablement and host/URL settings
- **THEN** ArgoCD sync renders the Istio `Gateway` and marketplace `HTTPRoute` resources from the marketplace chart

#### Scenario: Production ingress remains opt-in

- **WHEN** production manifests are rendered before production ingress enablement is chosen
- **THEN** production marketplace workloads do not expose an Istio ingress Gateway by accident

### Requirement: Staging hosted payment URL uses edge route

Staging value overlays SHALL set `HOSTED_PAYMENT_BASE_URL` to the Cloudflare-facing simulator HTTPS base URL when Istio ingress simulator routing is enabled.

#### Scenario: Staging simulator URL is public edge

- **WHEN** staging ingress with simulator routing is enabled
- **THEN** the web Deployment environment uses the public `https://` simulator hostname, not `http://payment-gateway-simulator:8097` cluster DNS alone and not `http://localhost:8097`

### Requirement: Staging Cloudflare Tunnel application

The repository SHALL include a staging Argo CD child Application that deploys in-cluster `cloudflared` for the marketplace edge.

#### Scenario: Staging root sync includes cloudflare-tunnel

- **WHEN** the staging root Application syncs from Git
- **THEN** Argo CD manages a child Application for the Cloudflare Tunnel connector in the `cloudflare-tunnel` namespace

#### Scenario: Tunnel token comes from External Secrets

- **WHEN** the cloudflare-tunnel chart syncs with External Secrets enabled
- **THEN** the tunnel token Secret is populated from Doppler via an ExternalSecret rather than committed to Git
