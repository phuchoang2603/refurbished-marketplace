## MODIFIED Requirements

### Requirement: Web keeps checkout scoped to one merchant group

Web MUST keep checkout scoped to one merchant group per submit. After validating the selected cart lines, prices, and shipping address, web MUST submit one durable intent to Checkout and present buyer-visible progress; it MUST redirect to hosted payment only when Checkout reports that the order is reserved and a session is ready.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group with a postal shipping address
- **THEN** web SHALL submit buyer and merchant ids, optional buyer email, authoritative total and named lines, shipping and return context, and the stable intent key to Checkout and SHALL show pending progress without directly calling Orders, Inventory, or Payment

#### Scenario: Merchant group exceeds products batch size at checkout

- **WHEN** a buyer submits more than 100 distinct product lines for a merchant group
- **THEN** web SHALL reject checkout with a clear browser-facing error before calling products batch or submitting an intent

#### Scenario: Reserve fails after place-order

- **WHEN** Checkout reports that inventory could not reserve the order
- **THEN** web SHALL show a browser-friendly unavailable-stock outcome and SHALL NOT redirect to hosted payment

#### Scenario: Checkout omits shipping address

- **WHEN** a buyer submits checkout without a usable postal address (line1, city, postal code, country)
- **THEN** web SHALL reject it before submitting an intent

#### Scenario: Buyer refreshes the pending page

- **WHEN** the buyer reloads the progress page while work is pending
- **THEN** web SHALL retrieve the same buyer-owned workflow and SHALL NOT create another order or session

### Requirement: Web re-validates cart products in one batch at checkout

Web MUST re-read selected merchant-group products through a single products batch call before submitting Checkout's intent and MUST use authoritative prices rather than cart snapshot prices.

#### Scenario: Buyer checks out one merchant group

- **WHEN** a buyer submits checkout for a merchant group with one or more lines within the products batch size limit
- **THEN** web SHALL batch-read those products once, SHALL fail closed on missing products or wrong merchant ids, and SHALL submit an authoritative price/quantity snapshot without removing cart lines

#### Scenario: Product missing at checkout re-validation

- **WHEN** batch lookup does not return a selected cart product
- **THEN** web SHALL reject the mutation without submitting a checkout intent

#### Scenario: Product merchant no longer matches checkout group

- **WHEN** a batch product's merchant id differs from the selected checkout group
- **THEN** web SHALL reject the mutation without submitting a checkout intent

### Requirement: Web checkout uses a stable per-submit intent

Web MUST send a stable per-submit buyer-scoped checkout intent identifier to Checkout so browser retries and double submits reuse one workflow and order.

#### Scenario: Checkout form is rendered

- **WHEN** web renders a merchant-group checkout action
- **THEN** it SHALL include a per-submit intent identifier reused by that same form

#### Scenario: Buyer retries the same checkout submit

- **WHEN** the buyer resubmits with the same intent after a transient failure
- **THEN** web SHALL submit the same key and render the original workflow instead of inventing a second order

### Requirement: Web defers cart remove until payment succeeds

Web MUST NOT require cart removal before checkout readiness. Once Checkout confirms a paid order after stock settlement, web MUST remove that order's paid merchant-group lines when a cart identity is available.

#### Scenario: Checkout POST succeeds through place-order and session create

- **WHEN** Checkout reports a ready hosted session for a reserved order
- **THEN** web SHALL redirect to hosted payment even if that merchant's cart lines remain

#### Scenario: Hosted payment callback reports success

- **WHEN** the buyer visits the confirmed order page and a cart identity is present
- **THEN** web SHALL remove the paid product ids with a multi-remove without affecting other merchants' cart items

### Requirement: Web accepts hosted payment outcome callbacks idempotently

Web MUST preserve the non-browser gateway callback route, forward verified outcomes to Payment without requiring browser authentication, and MUST NOT independently finalize Checkout, Orders, or Inventory. Repeated callbacks MUST be safe.

#### Scenario: Gateway posts a payment outcome callback

- **WHEN** a gateway callback reports an outcome for a hosted session
- **THEN** web SHALL forward it to Payment for durable outcome processing and SHALL NOT directly advance the checkout or mark an order paid

#### Scenario: Gateway retries a payment outcome callback

- **WHEN** the gateway repeats the same terminal callback
- **THEN** web SHALL acknowledge a safely processed duplicate without creating a new checkout, payment result, or order
