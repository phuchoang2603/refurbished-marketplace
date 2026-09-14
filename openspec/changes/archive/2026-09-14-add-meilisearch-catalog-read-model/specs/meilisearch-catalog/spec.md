## Purpose

Provide a GitOps-managed Meilisearch instance in the marketplace namespace as the storefront catalog search projection, not the listing or stock source of truth.

## ADDED Requirements

### Requirement: Meilisearch runs in ecommerce

The system SHALL deploy a single-node Meilisearch instance with persistent storage into the `ecommerce` namespace. Shop listing identity SHALL remain on Mongo. Live available and reserved quantity SHALL remain on inventory.

#### Scenario: Search instance is Ready

- **WHEN** the Meilisearch Application has synced and the pod is Ready
- **THEN** the HTTP search API is reachable in-cluster on the documented service port

#### Scenario: Search is not catalog source of truth

- **WHEN** Meilisearch is unavailable
- **THEN** CreateProduct, GetProductByID, GetProductsByIDs, GetStock, and ReserveStock SHALL still be able to succeed against Mongo and inventory

### Requirement: Search documents are catalog fields only

Indexed documents SHALL include listing identity, name, description, price, merchant, and created time. They SHALL NOT treat `initial_qty` or reservation outcomes as availability. They SHALL NOT store live available or reserved quantity.

#### Scenario: Create event is indexed

- **WHEN** a ProductCreated event is projected
- **THEN** a document for that product id exists with catalog fields and without live stock fields
