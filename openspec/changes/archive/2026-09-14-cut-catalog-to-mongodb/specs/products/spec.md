## MODIFIED Requirements

### Requirement: Products owns colocated stock state

The products service MUST own listing identity and catalog fields in MongoDB. It MUST NOT persist available or reserved quantity. It MUST NOT call the inventory service. Product reads SHALL return catalog fields only.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches a product by id for detail or cart stamping
- **THEN** the service SHALL return catalog fields from MongoDB and SHALL NOT load stock from inventory

#### Scenario: Product list is read

- **WHEN** a caller fetches a catalog product list before Meilisearch exists
- **THEN** the service SHALL return listing documents from MongoDB ordered by creation time and SHALL NOT read a SQL `products` table

### Requirement: Products creates listings with initial stock in one logical operation

The products service MUST persist the catalog document for seller-managed listing creation in MongoDB. It MUST NOT seed inventory. It SHALL require explicit non-negative initial quantity as creation intent and atomically persist a ProductCreated outbox document with the listing in one replica-set transaction. This intent SHALL NOT be exposed as live catalog stock.

#### Scenario: Seller-managed listing is created

- **WHEN** a trusted internal caller creates a product for an authenticated seller-managed listing
- **THEN** the service SHALL persist the catalog fields and creation outbox event atomically in MongoDB, then return the listing identity without waiting for inventory or search

#### Scenario: Seller-managed listing is created without explicit stock

- **WHEN** a caller attempts to create a seller-managed listing without explicit initial stock at the web boundary
- **THEN** web and products SHALL reject missing initial quantity instead of silently defaulting it; explicit zero SHALL be accepted and negative quantity rejected

### Requirement: Products supports batch lookup by IDs

The products service MUST expose a batch read that returns catalog rows for a set of product identifiers so checkout can re-validate many lines without N sequential single-get RPCs. The batch SHALL NOT join inventory. Catalog rows SHALL come from MongoDB.

#### Scenario: Multiple known IDs are requested

- **WHEN** a caller requests products by a non-empty list of product IDs within the service’s allowed batch size
- **THEN** the service SHALL return catalog records for every ID that exists in MongoDB, without stock summaries from the inventory ledger

#### Scenario: Some IDs are missing

- **WHEN** a batch request includes product IDs that do not exist in the catalog
- **THEN** the service SHALL omit those IDs from the response rather than failing the entire batch solely because some IDs are missing

#### Scenario: Batch size exceeds the limit

- **WHEN** a caller requests more product IDs than the configured maximum batch size
- **THEN** the service SHALL reject the request as invalid

### Requirement: Products durably publishes ProductCreated

Products SHALL write a ProductCreated outbox document in the same MongoDB replica-set transaction as catalog creation. CDC SHALL publish that document to `products.created`, keyed by product id. The event SHALL contain a stable event id, schema version, occurrence time, product id, initial product version, catalog name/description/price/merchant fields, and explicit initial quantity. Retries SHALL preserve event identity. Inventory and the future Meilisearch projector SHALL be able to consume the topic independently. Products SHALL NOT use a Postgres `products_outbox` table and SHALL NOT publish the creation event from the products process.

#### Scenario: Listing and event commit together

- **WHEN** CreateProduct succeeds
- **THEN** both listing and durable event SHALL exist in MongoDB, even if Kafka or a consumer is temporarily unavailable

#### Scenario: Outbox persistence fails

- **WHEN** either the listing write or outbox write fails
- **THEN** the transaction SHALL roll back both and CreateProduct SHALL fail

#### Scenario: Publication resumes after an outage

- **WHEN** Kafka publication resumes after a committed creation was delayed
- **THEN** the retained outbox event SHALL be published with its original identity without requiring the seller to recreate the listing

#### Scenario: Consumers progress independently

- **WHEN** one consumer group is delayed or replays ProductCreated
- **THEN** that group's progress SHALL NOT acknowledge or advance another group's consumption
