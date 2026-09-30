## MODIFIED Requirements

### Requirement: Cloudflare Tunnel is the public front door

Talos marketplace edges SHALL assume Cloudflare Tunnel as the public HTTPS front door and the Cilium Gateway as the HTTP origin. The talos-proxmox platform repository SHALL deploy the in-cluster `cloudflared` connector through Argo CD. This repository SHALL NOT require a marketplace TLS certificate on the Cilium Gateway for this path. The origin URL SHALL be `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80` without requiring the L2 announcement VIP for the tunnel.

#### Scenario: Origin is HTTP behind Cloudflare

- **WHEN** ingress is enabled for Cloudflare Tunnel access
- **THEN** the Cilium Gateway listens for HTTP from the in-cluster tunnel connector and does not require a marketplace TLS Secret for browser access

#### Scenario: Public hostnames match route hostnames

- **WHEN** Cloudflare Public Hostnames are configured for web and simulator
- **THEN** those hostnames match the Gateway/HTTPRoute hostname values and the origin URL uses the Cilium Gateway Service DNS (not `ecommerce-ingress-istio`)

#### Scenario: cloudflared is GitOps-managed

- **WHEN** `dev-root` or `prod-root` syncs from Git
- **THEN** the talos-proxmox platform root manages the `cloudflare-tunnel` Application that runs `cloudflared` with its bootstrap-provisioned token Secret
