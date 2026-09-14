## MODIFIED Requirements

### Requirement: Marketplace release owns databases

CNPG Clusters for marketplace services that still use Postgres SHALL be resources of the Argo-managed marketplace Helm release. The marketplace release SHALL NOT deploy a `products-db` Cluster after catalog cutover. The repository SHALL NOT apply remaining databases out-of-band to protect them from `tilt down`.

#### Scenario: Clusters sync with the chart

- **WHEN** the marketplace Application syncs
- **THEN** CNPG Cluster objects for remaining Postgres services are applied from the chart templates and no products Cluster is rendered

### Requirement: Product creation event transport

The GitOps configuration SHALL retain the `products.created` topic and SHALL deploy a Debezium MongoDB outbox connector against the catalog outbox collection, using Mongo credentials and the same EventRouter identity/key/payload/tracing mapping as other outbox connectors. It SHALL NOT deploy a Postgres products-outbox CDC connector. Inventory SHALL continue to subscribe to creation events with its own consumer group.

#### Scenario: Creation transport syncs

- **WHEN** the Kafka and marketplace Applications sync
- **THEN** the products creation topic, Mongo outbox connector, and inventory subscription SHALL be configured without a Postgres `products-outbox` connector and without replacing the inventory reservation event transport
