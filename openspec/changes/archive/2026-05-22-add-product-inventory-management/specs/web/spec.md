## ADDED Requirements

### Requirement: Web supports authenticated seller product creation

The web service MUST expose protected browser routes that let an authenticated user create a product with explicit initial stock through HTML form submission.

#### Scenario: Authenticated seller opens the create-product page

- **WHEN** an authenticated browser requests the seller product creation page
- **THEN** the web service SHALL render a server-side HTML page for that form

#### Scenario: Unauthenticated browser requests a seller management route

- **WHEN** an unauthenticated browser requests a protected seller product management route
- **THEN** the web service SHALL reject the request through the existing protected-route browser response instead of invoking downstream create operations

#### Scenario: Authenticated seller submits a product form

- **WHEN** an authenticated browser submits valid product details and an initial quantity
- **THEN** the web service SHALL call the products service to create the catalog record and initial stock through the unified catalog boundary before returning a usable success response

#### Scenario: Seller creation dependencies are unavailable

- **WHEN** a seller product creation request reaches the web service and a required downstream service is unavailable
- **THEN** the web service SHALL return a browser-friendly error response that matches the current popup-or-fragment mutation conventions

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart
- **THEN** the web service SHALL create one order for only that merchant's items and leave items from other merchants in the cart

### Requirement: Web maps seller identity to merchant ownership in v1

The web service MUST use the authenticated user's ID as the `merchant_id` value when creating seller-owned products until a separate merchant identity model exists.

#### Scenario: Seller creates a product in v1

- **WHEN** the web service handles product creation for an authenticated user
- **THEN** it SHALL submit that user's ID as the `merchant_id` in the product creation request

#### Scenario: Seller submits product creation without explicit stock

- **WHEN** an authenticated browser submits product creation without an explicit initial stock value
- **THEN** the web service SHALL reject the request using the current browser-friendly validation or error response path instead of relying on implicit defaults
