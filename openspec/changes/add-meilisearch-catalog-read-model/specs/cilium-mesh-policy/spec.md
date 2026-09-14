## ADDED Requirements

### Requirement: Products may reach Meilisearch without mesh mTLS

Meilisearch HTTP in `ecommerce` SHALL accept traffic from the `products` identity (and kubelet health checks). Those ingress rules SHALL NOT require Cilium mutual authentication. When mesh policy enforce is true, unknown identities SHALL be denied.

#### Scenario: Products can query search

- **WHEN** a client with the `products` Cilium identity dials Meilisearch HTTP after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Unknown identity is denied

- **WHEN** a pod that is not products (and is not kubelet) dials Meilisearch HTTP after enforce is true
- **THEN** Cilium policy drops the flow
