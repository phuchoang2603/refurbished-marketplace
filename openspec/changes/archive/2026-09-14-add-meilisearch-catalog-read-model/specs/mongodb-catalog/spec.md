## MODIFIED Requirements

### Requirement: Products catalog traffic uses the replica set

Authenticated products SHALL write catalog documents and serve GetProductByID and GetProductsByIDs from the MongoDB Community replica set in `ecommerce`. Shop create, detail, and checkout batch SHALL fail if Mongo is unavailable rather than falling back to Postgres. Storefront and seller lists SHALL NOT require a Mongo listing scan.

#### Scenario: Listing create requires Mongo

- **WHEN** products handles CreateProduct after this cutover
- **THEN** it SHALL persist the listing on the replica set and SHALL NOT insert a SQL `products` row

#### Scenario: Catalog reads require Mongo

- **WHEN** products handles GetProductByID or GetProductsByIDs
- **THEN** it SHALL load documents from the replica set

#### Scenario: Outbox change streams are available to CDC

- **WHEN** the catalog outbox collection exists on the replica set
- **THEN** an authenticated CDC client SHALL be able to watch inserts without converting standalone mongod to a replica set
