## MODIFIED Requirements

### Requirement: Authorization between marketplace identities

The system SHALL allow only documented caller identities to reach each marketplace HTTP/gRPC Service. When `meshPolicy.enforce` is true, CNPs SHALL set `enableDefaultDeny.ingress: true` so unknown callers are dropped.

#### Scenario: Unknown identity is denied

- **WHEN** a pod that is not an allowed caller sends traffic to a protected gRPC Service
- **THEN** Cilium policy drops the flow and the call fails

#### Scenario: Documented callers succeed

- **WHEN** `web` calls `orders` (or other documented pairs) after policies are enforced
- **THEN** checkout and related flows continue to succeed

#### Scenario: Orders reserves stock on products

- **WHEN** `orders` calls `products` gRPC to reserve stock for place-order after policies are enforced
- **THEN** Cilium policy SHALL allow that documented pair so checkout can hold stock before hosted payment
