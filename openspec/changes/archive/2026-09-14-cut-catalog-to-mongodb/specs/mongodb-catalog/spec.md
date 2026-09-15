## ADDED Requirements

### Requirement: Products catalog traffic uses the replica set

Authenticated products SHALL read and write catalog documents on the MongoDB Community replica set in `ecommerce`. Shop create, detail, batch, and seller list SHALL fail if Mongo is unavailable rather than falling back to Postgres.

#### Scenario: Listing create requires Mongo

- **WHEN** products handles CreateProduct after this cutover
- **THEN** it SHALL persist the listing on the replica set and SHALL NOT insert a SQL `products` row

#### Scenario: Catalog reads require Mongo

- **WHEN** products handles GetProductByID, GetProductsByIDs, or ListProducts after this cutover
- **THEN** it SHALL load documents from the replica set

#### Scenario: Outbox change streams are available to CDC

- **WHEN** the catalog outbox collection exists on the replica set
- **THEN** an authenticated CDC client SHALL be able to watch inserts without converting standalone mongod to a replica set

## REMOVED Requirements

### Requirement: Shop does not depend on Mongo after deploy

**Reason**: Catalog source of truth moves to Mongo in this change; unused `MONGO_ADDR` is not sufficient.
**Migration**: Products loads Mongo credentials and uses a driver against the existing replica set. Shop catalog RPCs depend on Mongo health.
