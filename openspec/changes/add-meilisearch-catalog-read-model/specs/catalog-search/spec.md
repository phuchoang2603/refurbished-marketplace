## ADDED Requirements

### Requirement: Catalog search read model uses Meilisearch

The platform MUST provide a Meilisearch-backed catalog search read model for storefront-oriented product listing and text search. PostgreSQL remains the source of truth for catalog and stock writes; Meilisearch MUST NOT accept authoritative catalog writes from sellers or the web edge.

#### Scenario: Search index is the list/search store

- **WHEN** a storefront-oriented catalog list or text search is executed
- **THEN** results SHALL be served from the Meilisearch catalog index rather than a PostgreSQL `LIMIT`/`OFFSET` catalog scan

#### Scenario: Writes do not target Meilisearch

- **WHEN** a seller or internal caller creates or updates catalog data through the products write API
- **THEN** the system SHALL persist the authoritative record in PostgreSQL and SHALL project into Meilisearch asynchronously rather than treating Meilisearch as the write store

### Requirement: Catalog documents are projected from the write model

The system MUST project catalog documents into Meilisearch from PostgreSQL-backed catalog change events (outbox → messaging → projector) so that creating a product results in a corresponding searchable document within the documented eventual-consistency window.

#### Scenario: Product create is projected

- **WHEN** a product is successfully created in PostgreSQL with a catalog outbox event
- **THEN** the projector SHALL upsert a Meilisearch document for that product including at least product id, name, description, price, merchant id, available quantity at create time, and created-at

#### Scenario: Projection failure is retried

- **WHEN** the projector fails to upsert a document into Meilisearch
- **THEN** the system SHALL retry according to consumer retry semantics and SHALL NOT mark the PostgreSQL write as failed after commit

### Requirement: Catalog index settings are versioned

The repository MUST version Meilisearch index settings for the products catalog (searchable attributes, filterable attributes, and ranking rules) so deploys can apply a known configuration.

#### Scenario: Index settings applied

- **WHEN** the catalog search index is provisioned or the projector/startup applies settings
- **THEN** the products index SHALL use the versioned searchable and filterable attributes required for text search and at least merchant or in-stock filtering

### Requirement: Reindex from PostgreSQL is documented and runnable

The system MUST provide a documented procedure to rebuild the Meilisearch catalog index from the PostgreSQL write model for drift recovery.

#### Scenario: Rebuild restores search parity

- **WHEN** an operator runs the documented catalog reindex from PostgreSQL
- **THEN** Meilisearch SHALL contain documents that match the current catalog write-model snapshot for projected fields

### Requirement: Eventual consistency is documented

The system MUST document expected projection lag and failure handling between PostgreSQL catalog writes and Meilisearch visibility.

#### Scenario: Operators know consistency expectations

- **WHEN** an operator or developer reads the catalog search documentation
- **THEN** the docs SHALL state that list/search is eventually consistent with the write model and SHALL describe retry/rebuild behavior
