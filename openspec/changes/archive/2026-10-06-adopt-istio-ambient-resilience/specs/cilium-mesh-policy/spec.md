## REMOVED Requirements

### Requirement: Strict mTLS for marketplace east-west

**Reason**: Strict mTLS is provided by Istio ambient rather than Cilium/SPIRE.
**Migration**: None; use the new ambient security contract.

#### Scenario: Enrolled traffic

- **WHEN** marketplace workload traffic crosses enrolled namespaces
- **THEN** no Cilium mutual-authentication policy is rendered

### Requirement: Authorization between marketplace identities

**Reason**: Marketplace Cilium L3/L4 ingress allow-lists are no longer required.
**Migration**: None; strict Istio mTLS does not imply a replacement per-service allow-list.

#### Scenario: Fresh policy render

- **WHEN** the charts render
- **THEN** no Cilium ingress default-deny policy is created

### Requirement: Gateway timeouts and outlier detection

**Reason**: Cilium Gateway outlier behavior and the pre-saga blanket retry prohibition are superseded by route-specific Istio resilience.
**Migration**: None; use `istio-resilience` for timeout, retry, and outlier contracts.

#### Scenario: Idempotent submit

- **WHEN** a checkout RPC with an identical durable intent is retried
- **THEN** the old blanket no-retry rule does not apply

### Requirement: Policy rollback

**Reason**: Rolling back to a Cilium mesh is not part of fresh-cluster operation.
**Migration**: None; validate the new Istio policy directly.

#### Scenario: Networking fault

- **WHEN** an operator investigates mesh connectivity
- **THEN** no Cilium mutual-authentication flag is used to restore traffic

### Requirement: Products may reach Mongo without mesh mTLS

**Reason**: Products and Mongo now share Istio ambient mTLS.
**Migration**: None; use the new ambient security contract.

#### Scenario: Catalog database connection

- **WHEN** Products connects to Mongo
- **THEN** the old plaintext exception is not required

### Requirement: Kafka Connect may reach Mongo without mesh mTLS

**Reason**: Kafka Connect and Mongo are now in enrolled namespaces.
**Migration**: None; use the new ambient security contract.

#### Scenario: Catalog CDC

- **WHEN** Kafka Connect queries Mongo
- **THEN** the old Cilium plaintext exception is not required

### Requirement: Inventory gRPC identities

**Reason**: Cilium identities and its gRPC allow-list are retired; Inventory traffic receives ambient mTLS and service-level resilience.
**Migration**: None; use the new ambient and resilience contracts.

#### Scenario: Inventory call

- **WHEN** Web reads Inventory stock
- **THEN** the connection does not require a Cilium identity rule

### Requirement: Search may reach Meilisearch without mesh mTLS

**Reason**: Search and Meilisearch now share Istio ambient mTLS.
**Migration**: None; use the new ambient security contract.

#### Scenario: Search backend

- **WHEN** Search queries Meilisearch
- **THEN** the old plaintext exception is not required

### Requirement: Search gRPC identities

**Reason**: Cilium identities and its Search allow-list are retired.
**Migration**: None; use the new ambient and resilience contracts.

#### Scenario: Search call

- **WHEN** Web calls Search gRPC
- **THEN** the connection does not require a Cilium identity rule
