## MODIFIED Requirements

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants, and after a successful place-order for that group it MUST initiate a hosted payment session for that order and redirect the buyer to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart with a postal shipping address
- **THEN** the web service SHALL place one order for only that merchant's items using the submit intent key, call products `ReserveStock` for that order, leave items from other merchants in the cart, request a hosted payment session for the created order including nested buyer and merchant ids, optional buyer email from the access token, total cents, shipping address, and named line items, and redirect the browser to the hosted payment URL

#### Scenario: Merchant group exceeds products batch size at checkout

- **WHEN** a buyer submits checkout for a merchant group with more distinct product lines than the products batch lookup limit (100)
- **THEN** the web service SHALL reject checkout with a clear browser-facing error before calling products batch or placing an order

#### Scenario: Reserve fails after place-order

- **WHEN** products cannot reserve the merchant group after the order is created
- **THEN** the web service SHALL fail the order, return a browser-friendly error, and SHALL NOT redirect to hosted payment

#### Scenario: Checkout omits shipping address

- **WHEN** a buyer submits checkout without a usable postal shipping address (at least line1, city, postal code, and country)
- **THEN** the web service SHALL reject checkout with a browser-friendly error and SHALL NOT place an order or create a hosted payment session
