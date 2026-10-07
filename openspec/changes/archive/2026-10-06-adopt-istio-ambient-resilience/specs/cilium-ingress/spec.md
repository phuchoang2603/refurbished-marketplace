## REMOVED Requirements

### Requirement: GitOps-managed Cilium edge gateway

**Reason**: Istio now owns application ingress.
**Migration**: None; use the new Istio Gateway contract on fresh clusters.

#### Scenario: Fresh ingress

- **WHEN** marketplace ingress is enabled
- **THEN** no `cilium` GatewayClass is required

### Requirement: Browser traffic reaches web through Cilium Gateway

**Reason**: Web traffic now traverses the Istio Gateway.
**Migration**: None; use the new Istio ingress contract.

#### Scenario: Web origin

- **WHEN** a shop hostname routes to Web
- **THEN** no Cilium Gateway handles that HTTP request

### Requirement: Hosted payment simulator edge exposure

**Reason**: Its browser route now uses the Istio Gateway.
**Migration**: None; use the new Istio ingress contract.

#### Scenario: Simulator origin

- **WHEN** a pay hostname routes to the simulator
- **THEN** no Cilium Gateway handles that HTTP request

### Requirement: Cloudflare Tunnel is the public front door

**Reason**: The Cilium-specific origin and tunnel assumptions are superseded; Cloudflare remains the public front door in the new capability.
**Migration**: None; bind the Istio-generated origin on fresh clusters.

#### Scenario: Tunnel origin

- **WHEN** the tunnel sends a request to the cluster
- **THEN** it does not depend on a Cilium-generated Gateway Service

### Requirement: TLS termination ownership is documented

**Reason**: The Cilium origin description is obsolete; TLS ownership is restated under Istio ingress.
**Migration**: None; use the new ingress documentation.

#### Scenario: Browser TLS

- **WHEN** an operator inspects the marketplace ingress contract
- **THEN** it does not prescribe a Cilium Gateway origin

### Requirement: Ingress rollback is documented

**Reason**: Cilium-specific rollback is irrelevant to the recreated clusters.
**Migration**: None; verify the Istio ingress configuration directly.

#### Scenario: Fresh-cluster troubleshooting

- **WHEN** a marketplace ingress route fails
- **THEN** remediation does not require restoring Cilium Gateway settings

### Requirement: Cilium CNI is cluster-owned

**Reason**: Ownership of Cilium networking is retained but separated from obsolete Cilium ingress requirements.
**Migration**: None; the new ambient networking capability retains platform CNI ownership.

#### Scenario: CNI owner

- **WHEN** the marketplace root syncs
- **THEN** it does not create a Cilium Gateway or manage the Cilium agent

### Requirement: Cluster deploy path is Argo CD and GHCR

**Reason**: This requirement belongs to the existing GitOps capability, not a removed Cilium ingress capability.
**Migration**: None; `argocd-gitops` retains the Argo CD and GHCR deployment contract.

#### Scenario: Deployment

- **WHEN** a contributor deploys marketplace images
- **THEN** Argo CD and GHCR remain the documented path without Cilium ingress
