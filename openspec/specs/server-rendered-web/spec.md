# Server Rendered Web

## Purpose

The server-rendered web capability defines how marketplace browser pages and interactive fragments are rendered by the web service.

## Requirements

### Requirement: Web renders server-side pages

The web service MUST render marketplace pages on the server as usable browser UI pages instead of requiring the browser to assemble them from JSON data, including auth-related pages and post-auth transitions.

#### Scenario: A page is requested directly

- **WHEN** a browser requests a top-level marketplace page
- **THEN** the web service SHALL return a complete HTML page for that route

#### Scenario: A page uses shared UI assets

- **WHEN** a browser requests a server-rendered page
- **THEN** the page SHALL include the shared marketplace shell plus the vendored DatastarUI-compatible styling, theme foundation, and component assets needed for a cohesive marketplace experience

#### Scenario: Auth flow completes

- **WHEN** a login, registration, or logout interaction succeeds
- **THEN** the web service SHALL render or redirect into a browser flow that is immediately useful for marketplace navigation

#### Scenario: Auth interruption resumes cart flow safely

- **WHEN** a guest browses products, adds items to the cart, and is interrupted by authentication at checkout
- **THEN** the post-login browser flow SHALL return the user to a usable cart or intended page state without exposing a token-debug view or silently replaying the original checkout mutation

### Requirement: Web supports Datastar fragment updates

The web service MUST return HTML fragments or Datastar SSE patch responses that Datastar can morph into existing DOM targets for interactive browser updates.

#### Scenario: A partial interaction is submitted

- **WHEN** a browser submits a Datastar-enabled interaction
- **THEN** the web service SHALL return HTML suitable for patching the targeted DOM element

#### Scenario: A fragment response is rendered

- **WHEN** the web service returns a fragment response
- **THEN** the response SHALL include markup with stable DOM IDs that match the target used by the interaction

#### Scenario: Router migration preserves fragment behavior

- **WHEN** the router implementation changes underneath the browser edge
- **THEN** Datastar-enabled routes SHALL preserve their HTML-first interaction model without introducing browser-facing JSON APIs

### Requirement: Web keeps internal composition over gRPC

The web service MUST continue to compose data from internal services over gRPC while rendering browser responses, and it MUST localize downstream read-path failures to the affected page, fragment, or interaction when a usable browser response can still be produced.

#### Scenario: A rendered page needs marketplace data

- **WHEN** a page requires data from internal domain services
- **THEN** the web service SHALL fetch that data through the existing gRPC clients before rendering HTML

#### Scenario: A page dependency is unavailable

- **WHEN** a browser page depends on a downstream service that is unavailable
- **THEN** the web service SHALL return a usable HTML page with inline localized unavailable-state content, partial content, or other feature-scoped fallback instead of failing the entire browser shell when that route can degrade safely

#### Scenario: A fragment dependency is unavailable

- **WHEN** a Datastar fragment or interactive browser update depends on a downstream service that is unavailable
- **THEN** the web service SHALL return localized browser feedback or a fallback fragment scoped to that feature instead of turning unrelated browser features unavailable

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

The web service MUST render cart items grouped by `merchant_id` and expose a separate checkout action per merchant group so checkout remains single-order per submit, and a successful checkout submit MUST continue the browser flow into a hosted payment redirect instead of stopping at a local payment-entry page.

#### Scenario: Cart contains items from multiple merchants

- **WHEN** the web service renders the cart page for a cart containing items from different merchants
- **THEN** the page SHALL group those items by merchant and render a merchant-scoped subtotal and checkout action for each group

#### Scenario: Merchant checkout action succeeds

- **WHEN** a buyer submits a merchant-scoped checkout action successfully
- **THEN** the browser flow SHALL continue into the hosted payment redirect for that created order instead of rendering a marketplace-hosted payment form

### Requirement: Web renders usable hosted payment return pages

The web service MUST render or redirect to a usable marketplace page when the buyer returns from the hosted payment gateway.

#### Scenario: Buyer returns after payment attempt

- **WHEN** the browser returns from the hosted payment gateway to the marketplace
- **THEN** the web service SHALL render or redirect to a complete HTML page that shows the buyer a usable post-payment state for the order
