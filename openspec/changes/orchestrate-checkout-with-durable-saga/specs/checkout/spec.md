## Purpose

Own the durable, merchant-scoped checkout workflow from buyer submission through stock reservation, hosted payment, stock settlement, and final order state, including compensation and reconciliation.

## ADDED Requirements

### Requirement: Checkout accepts a durable merchant-scoped intent

Checkout MUST persist a buyer-authenticated, merchant-scoped, validated order and payment snapshot with a buyer-scoped idempotency key before reporting acceptance. Acceptance MUST NOT imply that stock is reserved or the buyer can pay.

#### Scenario: Valid submit

- **WHEN** a buyer submits a valid merchant group, checkout intent, shipping address, and price/quantity snapshot
- **THEN** Checkout SHALL persist one workflow and return its identifier and pending status without waiting for order creation, inventory, or payment

#### Scenario: Duplicate or conflicting submit

- **WHEN** the same buyer retries an intent with the same facts or different facts
- **THEN** Checkout SHALL return the original workflow for identical facts or reject conflicting facts without creating another workflow or order

### Requirement: Checkout advances only on durable, correlated results

Checkout MUST command Orders to create an order, then Inventory to reserve all lines, then Payment to create a hosted session. It MUST expose a payment destination only after confirming stock is reserved and the session is ready; no downstream service SHALL infer workflow state from the mere passage of time.

#### Scenario: Reservation completes after buyer disconnects

- **WHEN** the buyer closes the browser while order creation and stock reservation are in progress
- **THEN** Checkout SHALL continue processing the workflow and show a ready payment destination on a later buyer-authenticated status read

#### Scenario: Stock is unavailable

- **WHEN** Inventory reports that full-order stock reservation failed
- **THEN** Checkout SHALL NOT request a hosted payment session and SHALL move the order to a failed state after confirming no stock remains held

#### Scenario: Payment session cannot be created

- **WHEN** a reserved order cannot obtain a hosted payment session
- **THEN** Checkout SHALL request stock release and SHALL NOT finalize the order as failed until Inventory confirms release

### Requirement: Checkout settles stock before final order status

Checkout MUST command Inventory to commit reserved stock after confirmed payment success, or release it after definite failure/cancellation, and MUST await the corresponding settlement result before finalizing Orders. The workflow MUST NOT issue a second payment attempt for the same order.

#### Scenario: Successful payment

- **WHEN** Payment reports a definitive success for the ready session
- **THEN** Checkout SHALL request stock commit and SHALL finalize the order as paid only after Inventory confirms commitment

#### Scenario: Definitive failed payment

- **WHEN** Payment reports a definitive failure or confirmed cancellation
- **THEN** Checkout SHALL request stock release and SHALL finalize the order as failed only after Inventory confirms release

### Requirement: Checkout survives replay and lost acknowledgements

Checkout MUST persist each processed result with its state change and outgoing command atomically, and MUST prevent duplicate, stale, or miscorrelated messages from initiating extra orders, reservations, sessions, stock movements, or order finalizations.

#### Scenario: Event is redelivered after commit

- **WHEN** a result is redelivered after its transition committed but its acknowledgement was lost
- **THEN** Checkout SHALL treat it as already processed and SHALL NOT emit a second effective command

#### Scenario: Old result arrives during a later state

- **WHEN** a valid but stale result for the same checkout arrives after the workflow advanced
- **THEN** Checkout SHALL retain the current state and SHALL NOT rewind or repeat side effects

### Requirement: Checkout uses durable deadlines and reconciles uncertain outcomes

Checkout MUST preserve order, reservation, and payment deadlines across restarts. When a command times out with an unknown outcome, Checkout MUST reconcile or issue a fencing cancellation before finalizing, and MUST surface unresolved financial outcomes for investigation rather than assuming failure.

#### Scenario: Reservation response is lost

- **WHEN** the reservation deadline passes without a definitive result
- **THEN** Checkout SHALL prevent a late reserve from creating a usable hold, obtain a definitive released/absent result, and only then finalize the order as failed

#### Scenario: Hosted payment expires while outcome is uncertain

- **WHEN** the payment deadline passes but a gateway capture may still have succeeded
- **THEN** Checkout SHALL keep the workflow non-terminal pending reconciliation or manual review and SHALL NOT blindly release stock or mark the order failed

#### Scenario: Success arrives after a terminal failure

- **WHEN** Payment reports a verified success after Checkout recorded a failed/expired attempt
- **THEN** Checkout SHALL flag a financial exception for reconciliation and SHALL NOT silently convert a failed or unreserved order to paid

### Requirement: Checkout restricts buyer-visible progress

Checkout MUST expose workflow progress and payment destination only to the authenticated buyer who owns the checkout, and MUST distinguish preparing, ready-to-pay, settling, compensated, completed, and needs-review states.

#### Scenario: Another buyer requests progress

- **WHEN** an authenticated buyer requests a checkout belonging to a different buyer
- **THEN** Checkout SHALL deny access without exposing the hosted session or commerce snapshot
