## MODIFIED Requirements

### Requirement: Products owns colocated stock state

The products service MUST own listing identity and catalog fields in MongoDB. It MUST NOT persist available or reserved quantity. It MUST NOT call the inventory service. Product reads by id and batch SHALL return catalog fields only from MongoDB. Storefront and seller lists SHALL use the search service, not a Mongo listing scan and not ListProducts.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches a product by id for detail or cart stamping
- **THEN** the service SHALL return catalog fields from MongoDB and SHALL NOT load stock from inventory

#### Scenario: Product list is read

- **WHEN** a caller needs a catalog product list
- **THEN** products SHALL NOT expose ListProducts; the caller SHALL use search SearchProducts

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for catalog reads by id, batch reads, and listing creation. It MUST NOT expose ReserveStock. It MUST NOT expose ListProducts or SearchProducts.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching catalog product or not-found

#### Scenario: Reserve is requested over gRPC

- **WHEN** a documented internal caller requests reservation for an order on the products API
- **THEN** products SHALL NOT apply the reserve; reservation SHALL be served by inventory

#### Scenario: Reserve is requested on products

- **WHEN** a caller invokes reservation on the products API
- **THEN** the method is absent; reservation SHALL go to inventory

#### Scenario: ListProducts is requested

- **WHEN** a caller invokes ListProducts on the products API
- **THEN** the method is absent; list and search SHALL go to the search service

### Requirement: Products durably publishes ProductCreated

Products SHALL write a ProductCreated outbox document in the same MongoDB replica-set transaction as catalog creation. CDC SHALL publish that document to `products.created`, keyed by product id. The event SHALL contain a stable event id, schema version, occurrence time, product id, initial product version, catalog name/description/price/merchant fields, and explicit initial quantity. Retries SHALL preserve event identity. Inventory and search SHALL consume the topic in independent consumer groups. Products SHALL NOT use a Postgres `products_outbox` table and SHALL NOT publish the creation event from the products process.

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
