# Search

## Purpose

Own the storefront catalog read model: consume ProductCreated into Meilisearch and serve SearchProducts. This service does not persist listings or stock.

## Requirements

### Requirement: Search projects ProductCreated into Meilisearch

Search SHALL consume ProductCreated on `products.created` in a consumer group that is not inventory's creation group and is not inventory's reservation/payment group. Successful consumption SHALL upsert the listing's catalog fields into Meilisearch. Projection lag or Meilisearch errors SHALL NOT stall inventory or products. CreateProduct SHALL NOT wait for the index upsert. Search SHALL NOT rebuild the index from Mongo.

#### Scenario: Creation appears in the index

- **WHEN** ProductCreated is published and the search consumer processes it
- **THEN** SearchProducts SHALL be able to return that listing

#### Scenario: Inventory consumption is isolated

- **WHEN** search projection fails or lags
- **THEN** inventory's ProductCreated and reservation consumers SHALL continue independently

#### Scenario: Products write path is isolated

- **WHEN** Meilisearch is down
- **THEN** CreateProduct SHALL still persist the listing and outbox in Mongo

### Requirement: Search exposes SearchProducts

The search service MUST expose SearchProducts over gRPC. Search SHALL query Meilisearch using optional text, optional merchant filter, and offset/limit. An empty text query SHALL return a browsable page of listings ordered by created time descending. A non-empty text query SHALL rank by Meilisearch relevance and SHALL NOT apply a recency sort. Results SHALL be catalog fields only and SHALL NOT include live stock. Search SHALL fail when the projection is unavailable rather than scanning Mongo.

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
- **THEN** the service SHALL return an unavailable error and SHALL NOT answer the list from Mongo
