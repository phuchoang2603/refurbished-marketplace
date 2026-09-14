## MODIFIED Requirements

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
- **THEN** the web service SHALL call products with catalog details and explicit initial quantity, and after the listing and creation event commit SHALL report listing creation with availability processing; it SHALL NOT call EnsureStock or compensating-delete the listing

#### Scenario: Seller creation dependencies are unavailable

- **WHEN** a seller product creation request reaches the web service and products cannot durably persist the listing and creation event
- **THEN** the web service SHALL return a browser-friendly error response that matches the current popup-or-fragment mutation conventions

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants, and after a successful place-order for that group it MUST initiate a hosted payment session for that order and redirect the buyer to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart with a postal shipping address
- **THEN** the web service SHALL place one order for only that merchant's items using the submit intent key, call inventory `ReserveStock` for that order, leave items from other merchants in the cart, request a hosted payment session for the created order including nested buyer and merchant ids, optional buyer email from the access token, total cents, shipping address, and named line items, and redirect the browser to the hosted payment URL

#### Scenario: Merchant group exceeds products batch size at checkout

- **WHEN** a buyer submits checkout for a merchant group with more distinct product lines than the products batch lookup limit (100)
- **THEN** the web service SHALL reject checkout with a clear browser-facing error before calling products batch or placing an order

#### Scenario: Reserve fails after place-order

- **WHEN** inventory cannot reserve the merchant group after the order is created
- **THEN** the web service SHALL fail the order, return a browser-friendly error, and SHALL NOT redirect to hosted payment

#### Scenario: Checkout omits shipping address

- **WHEN** a buyer submits checkout without a usable postal shipping address (at least line1, city, postal code, and country)
- **THEN** the web service SHALL reject checkout with a browser-friendly error and SHALL NOT place an order or create a hosted payment session

### Requirement: Web stamps cart line product snapshots

The web service MUST write product name and unit price onto the cart item from the PDP page snapshot (form or equivalent fields already shown to the buyer). Subsequent cart reads MUST NOT require per-line product hydration. Add-to-cart MUST NOT call GetProductByID or inventory solely to build that stamp. Checkout MUST still re-validate catalog prices via products batch.

#### Scenario: Buyer adds a product to the cart

- **WHEN** a browser adds a cart item with product_id, merchant_id, quantity, and the listing snapshot from the product page
- **THEN** the web service SHALL call cart add with those snapshot name and unit price fields and SHALL NOT call products GetProductByID solely to hydrate the stamp

#### Scenario: Buyer changes cart line quantity

- **WHEN** a browser sets quantity for an existing cart line (quantity greater than zero)
- **THEN** the web service SHALL reuse the stored cart snapshot (or the submitted snapshot) and SHALL NOT refresh live stock from inventory

#### Scenario: Cart becomes empty after remove or quantity zero

- **WHEN** a cart mutation (remove or set quantity to zero) leaves the cart with no items
- **THEN** the web service SHALL clear the browser `cart_id` cookie so the next cart action does not keep an empty cart document identity

## ADDED Requirements

### Requirement: Web distinguishes pending stock from unavailable stock

Web SHALL allow listing creation to succeed independently of inventory consumption. Until a read model exists, web MAY call GetStock once on PDP. A missing row SHALL be shown as availability processing, a read failure as availability unavailable, and an existing zero-quantity row as out of stock. Purchase controls SHALL be disabled while availability is pending or unavailable.

#### Scenario: Inventory has not consumed creation

- **WHEN** a newly created listing exists but its stock row is not found
- **THEN** web SHALL display pending availability without inventing zero stock or deleting the listing

#### Scenario: Stock read fails

- **WHEN** inventory cannot serve the PDP stock read
- **THEN** web SHALL display unavailable availability without claiming the product is out of stock

#### Scenario: Checkout arrives before stock initialization

- **WHEN** a buyer attempts checkout before the inventory row exists
- **THEN** synchronous ReserveStock SHALL fail, web SHALL fail the created order, and web SHALL NOT initiate hosted payment

#### Scenario: Inventory consumer is delayed during creation

- **WHEN** products commits the listing and ProductCreated while inventory consumption is delayed
- **THEN** web SHALL report listing creation without waiting for stock initialization or invoking compensation
