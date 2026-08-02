## ADDED Requirements

### Requirement: Web catalog browse and search use Meilisearch-backed APIs

The web service MUST load public catalog browse/search pages through the products service storefront list/search RPCs that read the Meilisearch catalog read model, including support for a text query parameter when search is exposed in the UI.

#### Scenario: Public catalog page lists from search API

- **WHEN** an unauthenticated browser client opens the public catalog listing page
- **THEN** the web service SHALL call the Meilisearch-backed products list/search API rather than relying on a PostgreSQL-only list path

#### Scenario: Catalog search query is forwarded

- **WHEN** a browser client provides a catalog search query string
- **THEN** the web service SHALL forward that query to the products search API and render the returned results

### Requirement: Seller product list uses server-side merchant scope

The web service MUST request seller-owned product listings using a merchant-scoped products list/search API (derived from the authenticated user as `merchant_id` in v1) instead of fetching a global catalog page and filtering merchant ownership only in the web process.

#### Scenario: Seller list is merchant-scoped

- **WHEN** an authenticated seller opens their product list page
- **THEN** the web service SHALL call the products list/search API with that seller’s `merchant_id` filter and SHALL NOT depend on client-side-only merchant filtering of a global catalog page for correctness
