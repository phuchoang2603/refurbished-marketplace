## MODIFIED Requirements

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants, and after a successful place-order for that group it MUST initiate a hosted payment session for that order and redirect the buyer to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart
- **THEN** the web service SHALL place one order for only that merchant's items using the submit intent key, call products `ReserveStock` for that order, leave items from other merchants in the cart, request a hosted payment session for the created order including merchant, total cents, optional shipping when present, and optional line items, and redirect the browser to the hosted payment URL

#### Scenario: Merchant group exceeds products batch size at checkout

- **WHEN** a buyer submits checkout for a merchant group with more distinct product lines than the products batch lookup limit (100)
- **THEN** the web service SHALL reject checkout with a clear browser-facing error before calling products batch or placing an order

#### Scenario: Reserve fails after place-order

- **WHEN** products cannot reserve the merchant group after the order is created
- **THEN** the web service SHALL fail the order, return a browser-friendly error, and SHALL NOT redirect to hosted payment

### Requirement: Web handles hosted payment return paths safely

The web service MUST provide browser routes that let a buyer return from the hosted payment gateway into a usable marketplace flow without replaying checkout or opening a new payment session for that order.

#### Scenario: Buyer returns after hosted payment completion

- **WHEN** the hosted payment gateway redirects the browser back after a successful, failed, or expired payment attempt
- **THEN** the web service SHALL redirect or render the buyer into a usable marketplace page for that order without issuing another checkout mutation or hosted-session create

## REMOVED Requirements

### Requirement: Web can resume hosted payment for a pending reserved order

**Reason:** Hosted payment is one-shot per order. Resume would mint or refresh a session on the same `order_id`.

**Migration:** Remove `POST /orders/{id}/pay` and any Resume payment control. After FAILED or EXPIRED, the buyer checks out again (new order).
