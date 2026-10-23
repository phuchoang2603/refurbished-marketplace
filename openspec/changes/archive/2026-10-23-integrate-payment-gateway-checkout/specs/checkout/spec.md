## MODIFIED Requirements

### Requirement: Checkout accepts a durable merchant-scoped intent

Checkout MUST persist a buyer-authenticated, merchant-scoped, validated order and payment snapshot with a buyer-scoped idempotency key before reporting acceptance. The snapshot MUST include each line's category and condition and the seller and buyer account creation times read from users. Acceptance MUST NOT imply that stock is reserved or the buyer can pay.

#### Scenario: Valid submit

- **WHEN** a buyer submits a valid merchant group, checkout intent, shipping address, and price/quantity snapshot
- **THEN** Checkout SHALL persist one workflow and return its identifier and pending status without waiting for order creation, inventory, or payment

#### Scenario: Duplicate or conflicting submit

- **WHEN** the same buyer retries an intent with the same facts or different facts
- **THEN** Checkout SHALL return the original workflow for identical facts or reject conflicting facts without creating another workflow or order

### Requirement: Checkout advances only on durable, correlated results

Checkout MUST command Orders to create an order, then Inventory to reserve all lines, then create a payment at the external payment gateway. It MUST expose the gateway's hosted checkout URL only after confirming stock is reserved and the gateway payment is recorded; no downstream service SHALL infer workflow state from the mere passage of time.

#### Scenario: Reservation completes after buyer disconnects

- **WHEN** the buyer closes the browser while order creation and stock reservation are in progress
- **THEN** Checkout SHALL continue processing the workflow and show a ready payment destination on a later buyer-authenticated status read

#### Scenario: Stock is unavailable

- **WHEN** Inventory reports that full-order stock reservation failed
- **THEN** Checkout SHALL NOT create a gateway payment and SHALL move the order to a failed state after confirming no stock remains held

#### Scenario: Payment session cannot be created

- **WHEN** a reserved order cannot obtain a gateway payment before its preparation deadline
- **THEN** Checkout SHALL request stock release and SHALL NOT finalize the order as failed until Inventory confirms release

### Requirement: Checkout settles stock before final order status

Checkout MUST command Inventory to commit reserved stock after verified gateway payment success, or release it after the gateway payment expired, and MUST await the corresponding settlement result before finalizing Orders. A declined attempt MUST NOT settle stock while the gateway payment is still open. The workflow MUST NOT create a second gateway payment for the same order.

#### Scenario: Successful payment

- **WHEN** a verified `payment.succeeded` webhook or deadline status read reports success for the order's gateway payment
- **THEN** Checkout SHALL request stock commit and SHALL finalize the order as paid only after Inventory confirms commitment

#### Scenario: Definitive failed payment

- **WHEN** a verified `payment.expired` webhook or deadline status read reports the gateway payment expired
- **THEN** Checkout SHALL request stock release and SHALL finalize the order as failed only after Inventory confirms release

### Requirement: Checkout uses durable deadlines and reconciles uncertain outcomes

Checkout MUST preserve order, reservation, and payment deadlines across restarts. When a command times out with an unknown outcome, Checkout MUST reconcile or issue a fencing cancellation before finalizing. For the payment step it MUST wait for the gateway's terminal webhook and read the gateway payment status at the deadline rather than cancelling, and MUST surface unresolved financial outcomes for investigation rather than assuming failure.

#### Scenario: Reservation response is lost

- **WHEN** the reservation deadline passes without a definitive result
- **THEN** Checkout SHALL prevent a late reserve from creating a usable hold, obtain a definitive released/absent result, and only then finalize the order as failed

#### Scenario: Hosted payment expires while outcome is uncertain

- **WHEN** the payment deadline passes and the gateway reports the payment neither succeeded nor expired
- **THEN** Checkout SHALL keep the workflow non-terminal, flag it for review and keep accepting gateway webhooks, without releasing stock or marking the order failed

#### Scenario: Success arrives after a terminal failure

- **WHEN** a gateway success arrives after Checkout recorded the payment as expired
- **THEN** Checkout SHALL flag a financial exception for reconciliation and SHALL NOT silently convert a failed or unreserved order to paid
