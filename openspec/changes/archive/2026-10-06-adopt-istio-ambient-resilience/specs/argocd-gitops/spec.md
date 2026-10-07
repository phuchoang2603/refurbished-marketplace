## REMOVED Requirements

### Requirement: Cilium ingress enablement on dest clusters

**Reason**: The platform disables Cilium Gateway API; Istio owns marketplace ingress.
**Migration**: None; use the new Istio ingress requirement on fresh clusters.

#### Scenario: Dev chart render

- **WHEN** dev marketplace ingress is enabled
- **THEN** it does not need a Cilium GatewayClass

## ADDED Requirements

### Requirement: Istio ingress enablement on dest clusters

The marketplace Application SHALL render Istio Gateway API resources when chart ingress is enabled. Chart defaults SHALL enable ingress for `shop-dev` / `pay-dev`; production SHALL use `shop` / `pay` via `values-prod.yaml`.

#### Scenario: Dev chart defaults enable ingress

- **WHEN** dev marketplace values use chart defaults in a Helm render
- **THEN** the output includes the Istio Gateway and marketplace HTTPRoutes without requiring a dev deployment

#### Scenario: Production overlay enables prod hosts

- **WHEN** prod-root applies `values-prod.yaml`
- **THEN** production marketplace workloads expose an Istio ingress Gateway for `shop` / `pay`

### Requirement: Namespace enrollment has one GitOps owner

The marketplace root SHALL declaratively create and enroll `ecommerce` and `kafka` in Istio ambient without competing namespace metadata owners, out-of-band labeling, or automatic enrollment of shared platform namespaces.

#### Scenario: Applications sync from empty state

- **WHEN** the production root starts against a fresh platform
- **THEN** both application namespaces have the ambient enrollment label before strict workload traffic depends on it

## MODIFIED Requirements

### Requirement: Kafka messaging namespace separation

The Kafka Application SHALL deploy Strimzi Kafka, Connect, and UI resources to a dedicated ambient-enrolled `kafka` namespace rather than `ecommerce`. Kafka's broker-level TLS behavior SHALL remain independent of Istio transport encryption, and Kafka Connect SHALL reach MongoDB in the enrolled `ecommerce` namespace.

#### Scenario: Kafka sync targets kafka namespace

- **WHEN** the Kafka Application syncs from Git
- **THEN** Kafka cluster resources are applied in an ambient-enrolled `kafka` namespace rather than `ecommerce`

#### Scenario: Marketplace reaches Kafka across namespaces

- **WHEN** marketplace services publish or consume messages
- **THEN** they use the Kafka bootstrap address in the `kafka` namespace DNS (for example `*.kafka.svc`)

#### Scenario: Connector reaches catalog database

- **WHEN** Kafka Connect synchronizes MongoDB changes
- **THEN** its cross-namespace database connection works under strict ambient mTLS

### Requirement: Marketplace consumes platform custom resources

Marketplace deployment SHALL retain CNPG Cluster, MongoDBCommunity, Kafka, KafkaNodePool, KafkaConnect, KafkaTopic, KafkaConnector, ExternalSecret, SecretStore, Gateway, HTTPRoute, PeerAuthentication, and DestinationRule resources needed by enabled workloads. It SHALL retain MongoDB workload service accounts/RBAC, Kafka connector secret-access RBAC, and schema initialization jobs. It SHALL NOT render CiliumNetworkPolicy or install shared operators, their CRDs, the CNI, or the platform observability stack.

#### Scenario: Consumer resources survive cleanup

- **WHEN** dev and prod application charts are rendered
- **THEN** enabled workload custom resources and supporting RBAC remain, no Cilium policy is rendered, and no shared operator installation or platform stack is rendered

#### Scenario: Empty databases

- **WHEN** marketplace workloads start against new empty databases
- **THEN** schema initialization still runs before dependent service operation
