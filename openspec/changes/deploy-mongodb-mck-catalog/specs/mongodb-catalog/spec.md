## Purpose

Provide a GitOps-managed MongoDB Community replica set in the marketplace namespace as the future catalog document store, without making shop traffic depend on Mongo yet.

## ADDED Requirements

### Requirement: MCK Community replica set in ecommerce

The system SHALL deploy a MongoDB Community replica set (not a standalone mongod) into the `ecommerce` namespace using MongoDB Controllers for Kubernetes Community custom resources. talos-dev SHALL use a single replica-set member unless an overlay raises the member count.

#### Scenario: Replica set is Running

- **WHEN** the MongoDB Community custom resource has been applied and reconciled
- **THEN** the resource reports a running replica set and `rs.status()` shows a replica set (not standalone)

#### Scenario: Change streams are possible

- **WHEN** a client authenticated to the replica set opens a change stream on an empty collection
- **THEN** the watch starts without requiring a standalone-to-replica-set conversion

### Requirement: Shop does not depend on Mongo after deploy

Marketplace HTTP and gRPC shop flows SHALL continue to succeed without reading or writing Mongo. Products MAY receive an unused Mongo host address in environment configuration.

#### Scenario: Browse and checkout still work

- **WHEN** Mongo is Healthy in `ecommerce` and products does not yet implement a Mongo driver
- **THEN** storefront browse, product detail, and checkout against Postgres continue to succeed

### Requirement: Community path only

The MongoDB deploy SHALL use the Community operator path. It SHALL NOT deploy Atlas, Ops Manager, or Cloud Manager, and SHALL NOT enable Enterprise continuous backup on the database resource.

#### Scenario: No Atlas or Ops Manager

- **WHEN** the GitOps applications for Mongo are synced
- **THEN** no Atlas Kubernetes Operator resources and no Ops Manager or Cloud Manager backup controllers are required for the replica set to become Ready
