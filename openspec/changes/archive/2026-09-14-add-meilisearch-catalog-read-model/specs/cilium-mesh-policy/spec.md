## ADDED Requirements

### Requirement: Search may reach Meilisearch without mesh mTLS

Meilisearch HTTP in `ecommerce` SHALL accept traffic from the `search` identity (and kubelet health checks). Those ingress rules SHALL NOT require Cilium mutual authentication. When mesh policy enforce is true, unknown identities SHALL be denied.

#### Scenario: Search can query Meilisearch

- **WHEN** a client with the `search` Cilium identity dials Meilisearch HTTP after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Unknown identity is denied

- **WHEN** a pod that is not search (and is not kubelet) dials Meilisearch HTTP after enforce is true
- **THEN** Cilium policy drops the flow

### Requirement: Search gRPC identities

Cilium policy SHALL allow `web` to dial search’s gRPC port. The products identity SHALL NOT be granted search ingress. Unknown identities SHALL be denied when mesh enforce is true. Search SHALL use the same Cilium mutual authentication mode as other enrolled marketplace gRPC services.

#### Scenario: Web can search

- **WHEN** web calls SearchProducts after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Unrelated pod is denied

- **WHEN** a pod that is not web dials search gRPC after enforce is true
- **THEN** Cilium drops the flow
