## ADDED Requirements

### Requirement: Inventory workload in marketplace chart

The marketplace Helm release SHALL deploy an `inventory` Service and Deployment in `ecommerce` with GHCR image tags matching other marketplace services.

#### Scenario: Inventory syncs with the chart

- **WHEN** the marketplace Application syncs
- **THEN** an inventory Deployment and Service exist in `ecommerce`

### Requirement: Product creation event transport

The GitOps configuration SHALL provision a `products.created` topic and a products outbox CDC connector using catalog database credentials, alongside the inventory outbox connector targeting inventory_db. Inventory SHALL subscribe to creation events with its own consumer group. The products connector SHALL have the required secret access and preserve event identity, product key, and tracing metadata.

#### Scenario: Creation transport syncs

- **WHEN** the Kafka and marketplace Applications sync
- **THEN** the products creation topic/connector and inventory subscription SHALL be configured without replacing the inventory reservation event transport
