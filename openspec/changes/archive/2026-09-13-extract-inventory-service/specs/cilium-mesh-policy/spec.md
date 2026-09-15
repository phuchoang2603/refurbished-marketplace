## ADDED Requirements

### Requirement: Inventory gRPC identities

Cilium policy SHALL allow `web` to dial inventory’s gRPC port. The products identity SHALL NOT be granted inventory ingress. Unknown identities SHALL be denied when mesh enforce is true. Inventory SHALL use the same Cilium mutual authentication mode as other enrolled marketplace gRPC services.

#### Scenario: Web can reserve

- **WHEN** web calls inventory ReserveStock after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Web can read stock

- **WHEN** web calls stock-read RPCs after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Unrelated pod is denied

- **WHEN** a pod that is not web dials inventory gRPC after enforce is true
- **THEN** Cilium drops the flow
