## MODIFIED Requirements

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants, and after creating that merchant-scoped order it MUST initiate a hosted payment session for that order and redirect the buyer to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart
- **THEN** the web service SHALL create one order for only that merchant's items, leave items from other merchants in the cart, request a hosted payment session for the created order, and redirect the browser to the hosted payment URL

## ADDED Requirements

### Requirement: Web handles hosted payment return paths safely

The web service MUST provide browser routes that let a buyer return from the hosted payment gateway into a usable marketplace flow without replaying checkout.

#### Scenario: Buyer returns after hosted payment completion

- **WHEN** the hosted payment gateway redirects the browser back after a successful or failed payment attempt
- **THEN** the web service SHALL redirect or render the buyer into a usable marketplace page for that order without issuing another checkout mutation

#### Scenario: Buyer returns after canceling hosted payment

- **WHEN** the hosted payment gateway redirects the browser back after the buyer cancels payment
- **THEN** the web service SHALL return the buyer to a usable order-related page without creating a second order or payment session

### Requirement: Web accepts hosted payment outcome callbacks idempotently

The web service MUST preserve a non-browser callback path for hosted payment outcomes and MUST handle repeated callbacks safely.

#### Scenario: Gateway posts a payment outcome callback

- **WHEN** the hosted payment gateway posts an order-level payment outcome callback
- **THEN** the web service SHALL process the outcome through the payment and order path without requiring browser authentication

#### Scenario: Gateway retries a payment outcome callback

- **WHEN** the hosted payment gateway repeats the same terminal callback for an order
- **THEN** the web service SHALL handle the duplicate safely without creating a second payment result or replaying checkout behavior
