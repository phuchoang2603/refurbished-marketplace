## ADDED Requirements

### Requirement: Web checkout uses a stable per-submit intent

The web service MUST send a stable per-submit checkout intent identifier on place-order so browser retries and double submits reuse one order.

#### Scenario: Checkout form is rendered

- **WHEN** the web service renders a merchant-group checkout action
- **THEN** it SHALL include a per-submit intent identifier that is reused if that same form is submitted again

#### Scenario: Buyer retries the same checkout submit

- **WHEN** the buyer resubmits checkout with the same intent identifier after a transient failure
- **THEN** the web service SHALL call place-order with that same key and SHALL NOT invent a second intent for that submit

### Requirement: Web defers cart remove until payment succeeds

The web service MUST NOT require cart multi-remove to succeed before redirecting to hosted payment. After a successful terminal payment callback for that order, it MUST remove the paid merchant-group lines.

#### Scenario: Checkout POST succeeds through place-order and session create

- **WHEN** place-order and hosted session creation succeed
- **THEN** the web service SHALL redirect to hosted payment even if cart lines for that merchant group are still present

#### Scenario: Hosted payment callback reports success

- **WHEN** the web service processes a successful terminal hosted-payment callback for an order
- **THEN** it SHALL remove the corresponding paid product IDs from the buyer cart with a multi-remove when a cart identity is available

### Requirement: Web can resume hosted payment for a pending reserved order

The web service MUST let a buyer obtain a hosted payment redirect for an existing unpaid reserved order without placing a second order.

#### Scenario: Buyer resumes payment from the order page

- **WHEN** a buyer requests to continue payment for their unpaid reserved order
- **THEN** the web service SHALL request a hosted session for that `order_id` and redirect to the hosted payment URL

## MODIFIED Requirements

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants, and after a successful place-order for that group it MUST initiate a hosted payment session for that order and redirect the buyer to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart
- **THEN** the web service SHALL place one order for only that merchant's items using the submit intent key, leave items from other merchants in the cart, request a hosted payment session for the created order, and redirect the browser to the hosted payment URL

#### Scenario: Merchant group exceeds products batch size at checkout

- **WHEN** a buyer submits checkout for a merchant group with more distinct product lines than the products batch lookup limit (100)
- **THEN** the web service SHALL reject checkout with a clear browser-facing error before calling products batch or placing an order

#### Scenario: Place-order fails because stock cannot be reserved

- **WHEN** place-order fails because products cannot reserve the merchant group
- **THEN** the web service SHALL return a browser-friendly error and SHALL NOT redirect to hosted payment

### Requirement: Web re-validates cart products in one batch at checkout

The web service MUST re-read selected merchant group products through a single batch products API before placing an order and MUST use those authoritative prices for order lines rather than cart snapshot prices.

#### Scenario: Buyer checks out one merchant group

- **WHEN** a buyer submits checkout for a merchant group containing one or more cart lines within the products batch size limit
- **THEN** the web service SHALL call products once with all selected product IDs for that group, SHALL fail closed if any selected product is missing or has a different merchant_id, SHALL place the order using batch-returned unit prices, and SHALL NOT require a cart multi-remove before the hosted payment redirect

#### Scenario: Product missing at checkout re-validation

- **WHEN** the batch product lookup does not return a product referenced by a selected cart line
- **THEN** the web service SHALL fail the checkout mutation without placing an order for that submit

#### Scenario: Product merchant no longer matches checkout group

- **WHEN** the batch product lookup returns a product whose merchant_id differs from the selected checkout merchant group
- **THEN** the web service SHALL fail the checkout mutation without placing an order for that submit
