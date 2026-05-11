## ADDED Requirements

### Requirement: Web UI provides a cohesive marketplace shell

The web service MUST provide a cohesive browser UI shell for marketplace pages, including shared navigation, typography, spacing, and responsive layout suitable for desktop and mobile browsers.

#### Scenario: A marketplace page is rendered

- **WHEN** a browser requests a marketplace page
- **THEN** the page SHALL render within the shared marketplace shell with navigation to catalog, cart, orders, and auth flows

#### Scenario: A page is viewed on a narrow screen

- **WHEN** a browser renders the web UI on a mobile-width viewport
- **THEN** the layout SHALL remain usable without horizontal scrolling for primary content and controls

#### Scenario: Shared primitives are rendered

- **WHEN** a page renders repeated UI primitives such as buttons, fields, cards, or empty states
- **THEN** the page SHALL use consistent copied DatastarUI-inspired server-rendered components wherever suitable

### Requirement: Web UI exposes catalog and product browsing

The web service MUST present catalog and product detail pages with user-friendly product information and actions.

#### Scenario: Catalog page is requested

- **WHEN** a browser requests the catalog route
- **THEN** the page SHALL display a browsable product list with names, prices, and links to product detail pages

#### Scenario: Product detail page is requested

- **WHEN** a browser requests a product detail route
- **THEN** the page SHALL display product details and an add-to-cart form

### Requirement: Web UI supports cart interaction

The web service MUST provide a cart UI that supports viewing product details for cart items, changing quantities, removing items, and receiving fragment updates after cart actions.

#### Scenario: A product is added to cart

- **WHEN** a browser submits the add-to-cart form
- **THEN** the web service SHALL return an HTML fragment or Datastar SSE patch suitable for updating the cart target

#### Scenario: Cart page is requested

- **WHEN** a browser requests the cart page
- **THEN** the cart rows SHALL include product details composed from the products service where available

#### Scenario: Cart quantity is changed

- **WHEN** a browser submits a quantity update from the cart page
- **THEN** the web service SHALL return an updated cart fragment or Datastar SSE patch

### Requirement: Web UI supports browser auth forms

The web service MUST provide browser-friendly auth forms and responses for login, logout, and account creation flows.

#### Scenario: Login page is requested

- **WHEN** a browser requests the login route
- **THEN** the page SHALL display a login form that submits form fields to the web service

#### Scenario: Login succeeds

- **WHEN** a browser submits valid login credentials
- **THEN** the web service SHALL persist browser auth tokens in HTTP cookies and return an HTML response for the authenticated browser flow

#### Scenario: Auth form submission fails

- **WHEN** a browser submits invalid auth form data
- **THEN** the web service SHALL return an HTML response that explains the error without exposing raw JSON
