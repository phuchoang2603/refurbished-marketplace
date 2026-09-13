## ADDED Requirements

### Requirement: Products may reach Mongo without mesh mTLS

Mongod endpoints in `ecommerce` SHALL accept TCP `27017` from the `products` identity and from kubelet health checks. Those ingress rules SHALL NOT require Cilium mutual authentication. When mesh policy enforce is true, unknown identities SHALL be denied.

#### Scenario: Products can connect

- **WHEN** a client with the `products` Cilium identity dials Mongo on port 27017 after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Unknown identity is denied

- **WHEN** a pod that is not `products` (and is not kubelet) dials Mongo on port 27017 after enforce is true
- **THEN** Cilium policy drops the flow and the probe fails

#### Scenario: Mongo is not enrolled in SPIRE mTLS

- **WHEN** products opens a Mongo driver or `mongosh` session to the replica set
- **THEN** the hop does not require Cilium `authentication.mode: required`
