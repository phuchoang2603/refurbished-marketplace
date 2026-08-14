## ADDED Requirements

### Requirement: CreateProduct emits a catalog outbox event

The products service MUST persist a catalog outbox event in the same database transaction as a successful `CreateProduct` write so downstream projection can update the Meilisearch read model without a dual-write from the API request path.

#### Scenario: Product create writes outbox atomically

- **WHEN** `CreateProduct` successfully inserts the product and initial inventory rows
- **THEN** the same transaction SHALL insert a catalog outbox record describing that product create (or equivalent catalog upsert payload)

#### Scenario: Failed create writes no outbox

- **WHEN** `CreateProduct` fails before commit
- **THEN** the service SHALL NOT leave a committed catalog outbox event for that failed create

### Requirement: Storefront list and search read the catalog search model

The products service MUST expose storefront-oriented catalog list/search RPCs that query the Meilisearch catalog read model, supporting text search and at least one filter (merchant id or in-stock / available quantity).

#### Scenario: Text by name substring

- **WHEN** a caller invokes catalog search with a non-empty query string
- **THEN** the service SHALL return matching products from Meilisearch using the configured searchable attributes

#### Scenario: Filter by merchant

- **WHEN** a caller invokes catalog search or list with a merchant filter
- **THEN** the service SHALL return only documents for that merchant from Meilisearch

#### Scenario: Browse without query

- **WHEN** a caller requests a storefront catalog page with an empty search query
- **THEN** the service SHALL return paginated catalog documents from Meilisearch rather than scanning PostgreSQL with `LIMIT`/`OFFSET` for that storefront path

## MODIFIED Requirements

### Requirement: Product list is read

The products service MUST support catalog product list reads for storefront-oriented flows via the Meilisearch-backed catalog search read model. Stock-exact detail and admin-oriented reads MAY remain on the PostgreSQL write model. Storefront list responses MAY omit perfectly fresh reserved/available quantities when quantity projection lags; detail reads remain authoritative for stock.

#### Scenario: Product list is read

- **WHEN** a caller fetches a storefront catalog product list
- **THEN** the service SHALL serve that list from the Meilisearch catalog read model and SHALL NOT require exact live stock quantities on every list row

#### Scenario: Detail remains stock-aware on the write model

- **WHEN** a caller fetches product data for a detail or admin-oriented stock-aware catalog flow
- **THEN** the service SHALL return product data from the unified PostgreSQL catalog boundary without requiring Meilisearch for correctness of stock
