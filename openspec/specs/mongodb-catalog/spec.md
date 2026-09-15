# MongoDB Catalog

## Purpose

Provide a GitOps-managed MongoDB Community replica set in the marketplace namespace as the future catalog document store, without making shop traffic depend on Mongo yet.

## Requirements

### Requirement: MCK Community replica set in ecommerce

The system SHALL deploy a MongoDB Community replica set (not a standalone mongod) into the `ecommerce` namespace using MongoDB Controllers for Kubernetes Community custom resources. talos-dev SHALL use a single replica-set member unless an overlay raises the member count.

#### Scenario: Replica set is Running

- **WHEN** the MongoDB Community custom resource has been applied and reconciled
- **THEN** the resource reports a running replica set and `rs.status()` shows a replica set (not standalone)

#### Scenario: Change streams are possible

- **WHEN** a client authenticated to the replica set opens a change stream on an empty collection
- **THEN** the watch starts without requiring a standalone-to-replica-set conversion

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

### Requirement: Community path only

The MongoDB deploy SHALL use the Community operator path. It SHALL NOT deploy Atlas, Ops Manager, or Cloud Manager, and SHALL NOT enable Enterprise continuous backup on the database resource.

#### Scenario: No Atlas or Ops Manager

- **WHEN** the GitOps applications for Mongo are synced
- **THEN** no Atlas Kubernetes Operator resources and no Ops Manager or Cloud Manager backup controllers are required for the replica set to become Ready
