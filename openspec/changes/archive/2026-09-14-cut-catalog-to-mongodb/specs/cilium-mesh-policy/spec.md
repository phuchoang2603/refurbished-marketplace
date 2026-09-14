## ADDED Requirements

### Requirement: Kafka Connect may reach Mongo without mesh mTLS

Mongod endpoints in `ecommerce` SHALL accept TCP `27017` from Kafka Connect as well as from the `products` identity and kubelet health checks. Those ingress rules SHALL NOT require Cilium mutual authentication. When mesh policy enforce is true, unknown identities SHALL be denied.

#### Scenario: Connect can watch the catalog outbox

- **WHEN** a Kafka Connect task with the documented Connect identity dials Mongo on port 27017 after policies are enforced
- **THEN** the connection is allowed

#### Scenario: Unknown identity is still denied

- **WHEN** a pod that is not products, kubelet, or Kafka Connect dials Mongo on port 27017 after enforce is true
- **THEN** Cilium policy drops the flow
