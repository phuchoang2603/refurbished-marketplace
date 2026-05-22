## ADDED Requirements

### Requirement: Web renders seller management pages

The web service MUST render seller-facing product management pages as complete server-side HTML experiences that match the marketplace's existing browser shell.

#### Scenario: Seller requests product management UI

- **WHEN** a browser requests a seller product management route
- **THEN** the web service SHALL return a complete HTML page with the product creation form and shared marketplace layout

#### Scenario: Seller page dependency is unavailable

- **WHEN** the seller product management page is requested
- **THEN** the web service SHALL render the static seller form shell without requiring a downstream read before the page loads

### Requirement: Web renders stock-aware product views

The web service MUST render product detail pages from stock-aware products data over internal gRPC so those pages reflect current availability instead of placeholder stock values.

#### Scenario: Product detail page is rendered

- **WHEN** the web service renders a product page for a stocked product
- **THEN** the page SHALL display stock information derived from the products service response

#### Scenario: Product list page is rendered

- **WHEN** the web service renders the product list page in this change
- **THEN** the page SHALL remain catalog-focused and SHALL NOT require exact per-item stock reads

### Requirement: Web renders merchant-grouped cart checkout

The web service MUST render cart items grouped by `merchant_id` and expose a separate checkout action per merchant group so checkout remains single-order per submit.

#### Scenario: Cart contains items from multiple merchants

- **WHEN** the web service renders the cart page for a cart containing items from different merchants
- **THEN** the page SHALL group those items by merchant and render a merchant-scoped subtotal and checkout action for each group
