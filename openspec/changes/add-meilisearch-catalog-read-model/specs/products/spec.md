## ADDED Requirements

### Requirement: Products searches the catalog projection

The products service MUST expose SearchProducts over gRPC. Search SHALL query the Meilisearch catalog projection using optional text, optional merchant filter, and offset/limit. An empty text query SHALL return a browsable page of listings. Results SHALL be catalog fields only and SHALL NOT include live stock. Search SHALL fail when the projection is unavailable rather than scanning Mongo as a list fallback.

#### Scenario: Empty query is browse

- **WHEN** a caller invokes SearchProducts with empty text, no merchant filter, and a valid limit
- **THEN** the service SHALL return catalog hits from the projection ordered for browse, without available or reserved quantity

#### Scenario: Text query matches listings

- **WHEN** a caller invokes SearchProducts with text that matches indexed name or description
- **THEN** the service SHALL return matching catalog hits from the projection

#### Scenario: Merchant filter is applied server-side

- **WHEN** a caller invokes SearchProducts with a merchant filter
- **THEN** every returned hit SHALL belong to that merchant

#### Scenario: Projection is down

- **WHEN** Meilisearch is unavailable during SearchProducts
- **THEN** the service SHALL return an unavailable error and SHALL NOT answer the list from a Mongo collection scan

### Requirement: Products projects ProductCreated into search

Products SHALL consume ProductCreated on `products.created` in a consumer group that is not inventory's creation group and is not inventory's reservation/payment group. Successful consumption SHALL upsert the listing's catalog fields into Meilisearch. Projection lag or Meilisearch errors SHALL NOT stall inventory consumers. CreateProduct SHALL NOT wait for the index upsert. The projector SHALL be able to rebuild the index from Mongo listings.

#### Scenario: Creation appears in the index

- **WHEN** ProductCreated is published and the search consumer processes it
- **THEN** SearchProducts SHALL be able to return that listing without a Mongo list scan

#### Scenario: Inventory consumption is isolated

- **WHEN** search projection fails or lags
- **THEN** inventory's ProductCreated and reservation consumers SHALL continue independently

#### Scenario: Rebuild restores the index

- **WHEN** operators rebuild search from Mongo
- **THEN** indexed documents SHALL match current listing documents for identity and catalog fields

## MODIFIED Requirements

### Requirement: Products owns colocated stock state

The products service MUST own listing identity and catalog fields in MongoDB. It MUST NOT persist available or reserved quantity. It MUST NOT call the inventory service. Product reads by id and batch SHALL return catalog fields only from MongoDB. Storefront and seller lists SHALL use SearchProducts, not a Mongo listing scan.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches a product by id for detail or cart stamping
- **THEN** the service SHALL return catalog fields from MongoDB and SHALL NOT load stock from inventory

#### Scenario: Product list is read

- **WHEN** a caller fetches a catalog product list
- **THEN** the service SHALL return SearchProducts hits from the catalog projection and SHALL NOT expose ListProducts

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for catalog reads by id, batch reads, SearchProducts, and listing creation. It MUST NOT expose ReserveStock. It MUST NOT expose ListProducts.

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
- **THEN** the method is absent; list and search SHALL go to SearchProducts

### Requirement: Products durably publishes ProductCreated

Products SHALL write a ProductCreated outbox document in the same MongoDB replica-set transaction as catalog creation. CDC SHALL publish that document to `products.created`, keyed by product id. The event SHALL contain a stable event id, schema version, occurrence time, product id, initial product version, catalog name/description/price/merchant fields, and explicit initial quantity. Retries SHALL preserve event identity. Inventory and the search projector SHALL consume the topic in independent consumer groups. Products SHALL NOT use a Postgres `products_outbox` table and SHALL NOT publish the creation event from the products process.

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
