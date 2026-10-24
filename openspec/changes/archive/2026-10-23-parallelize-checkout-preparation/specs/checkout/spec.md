## MODIFIED Requirements

### Requirement: Checkout accepts a durable merchant-scoped intent

Checkout MUST persist a buyer-authenticated, merchant-scoped, validated order and payment snapshot with a buyer-scoped idempotency key before reporting acceptance. The snapshot MUST include each line's category and condition and the seller and buyer account creation times read from users. Acceptance MUST atomically persist one shared preparation command containing the frozen order snapshot and the identifiers needed by both Orders and Inventory. Acceptance MUST NOT imply that stock is reserved or the buyer can pay.

#### Scenario: Valid submit

- **WHEN** a buyer submits a valid merchant group, checkout intent, shipping address, and price/quantity snapshot
- **THEN** Checkout SHALL persist one workflow and return its identifier and pending status without waiting for order creation, inventory, or payment

#### Scenario: Duplicate or conflicting submit

- **WHEN** the same buyer retries an intent with the same facts or different facts
- **THEN** Checkout SHALL return the original workflow for identical facts or reject conflicting facts without creating another workflow or order

#### Scenario: Acceptance transaction fails

- **WHEN** persistence of the workflow or its shared preparation command fails
- **THEN** Checkout SHALL commit neither and SHALL NOT report acceptance

### Requirement: Checkout advances only on durable, correlated results

Checkout MUST publish one shared preparation command so Orders can create the order and Inventory can reserve all lines independently in distinct consumer groups. Checkout MUST durably join the correlated successful results from both services before creating a payment at the external payment gateway. It MUST expose the gateway's hosted checkout URL only after confirming stock is reserved and the gateway payment is recorded; no downstream service SHALL infer workflow state from the mere passage of time.

#### Scenario: Reservation completes after buyer disconnects

- **WHEN** the buyer closes the browser while order creation and stock reservation are in progress
- **THEN** Checkout SHALL continue processing the workflow and show a ready payment destination on a later buyer-authenticated status read

#### Scenario: Stock is unavailable

- **WHEN** Inventory reports that full-order stock reservation failed
- **THEN** Checkout SHALL NOT create a gateway payment and SHALL move the order to a failed state after confirming no stock remains held

#### Scenario: Payment session cannot be created

- **WHEN** a reserved order cannot obtain a gateway payment before its preparation deadline
- **THEN** Checkout SHALL request stock release and SHALL NOT finalize the order as failed until Inventory confirms release

#### Scenario: Preparation results arrive in either order

- **WHEN** one participant reports successful preparation while the other result is pending
- **THEN** Checkout SHALL persist that branch outcome, remain preparing, and SHALL NOT create a gateway payment or publish a second initial branch command

#### Scenario: Both preparation branches succeed concurrently

- **WHEN** both correlated successful preparation results are processed concurrently
- **THEN** Checkout SHALL persist both outcomes and schedule exactly one effective gateway creation operation

#### Scenario: Orders rejects after stock reservation

- **WHEN** Orders reports a definitive preparation rejection and Inventory has reserved stock or its result is unknown
- **THEN** Checkout SHALL close preparation, request stock cancellation, await release or absence, and obtain failed order finalization or an acknowledged durable absence fence before marking the workflow failed

### Requirement: Checkout survives replay and lost acknowledgements

Checkout MUST persist each processed result with its state change and outgoing command atomically, and MUST prevent duplicate, stale, or miscorrelated messages from initiating extra orders, reservations, sessions, stock movements, or order finalizations.

#### Scenario: Event is redelivered after commit

- **WHEN** a result is redelivered after its transition committed but its acknowledgement was lost
- **THEN** Checkout SHALL treat it as already processed and SHALL NOT emit a second effective command

#### Scenario: Old result arrives during a later state

- **WHEN** a valid but stale result for the same checkout arrives after the workflow advanced
- **THEN** Checkout SHALL retain the current state and SHALL NOT rewind or repeat side effects

#### Scenario: Sibling result follows a committed result

- **WHEN** a valid result from one preparation branch arrives after the other branch updated the workflow
- **THEN** Checkout SHALL correlate it to the same preparation attempt independently of the updated workflow row version and SHALL NOT discard it as stale

#### Scenario: Conflicting branch result

- **WHEN** a participant reports an outcome contradictory to its already recorded definitive outcome
- **THEN** Checkout SHALL preserve the committed outcome, prevent extra side effects, and surface the conflict for reconciliation

### Requirement: Checkout uses durable deadlines and reconciles uncertain outcomes

Checkout MUST preserve the shared preparation recovery timeout and payment deadlines across restarts. The preparation timeout MUST make an open preparation eligible for recovery rather than impose a strict expiry on successful results. Checkout MUST serialize the successful join and the timeout worker's compensation decision so that whichever commits first determines the next step. Successful preparation results MAY complete the join after the timeout becomes due if preparation remains open. Once rejection or timeout compensation closes preparation, Checkout MUST NOT reopen it or create a payment from later preparation results. When recovery claims a timed-out command with an unknown outcome, Checkout MUST reconcile or issue a fencing cancellation before finalizing. For the payment step it MUST wait for the gateway's terminal webhook and read the gateway payment status at the deadline rather than cancelling, and MUST surface unresolved financial outcomes for investigation rather than assuming failure.

#### Scenario: Reservation response is lost

- **WHEN** the timeout worker claims preparation after its recovery timeout with the reservation outcome still unknown
- **THEN** Checkout SHALL prevent a late reserve from creating a usable hold, obtain a definitive released/absent result, and only then finalize the order as failed

#### Scenario: Hosted payment expires while outcome is uncertain

- **WHEN** the payment deadline passes and the gateway reports the payment neither succeeded nor expired
- **THEN** Checkout SHALL keep the workflow non-terminal, flag it for review and keep accepting gateway webhooks, without releasing stock or marking the order failed

#### Scenario: Success arrives after a terminal failure

- **WHEN** a gateway success arrives after Checkout recorded the payment as expired
- **THEN** Checkout SHALL flag a financial exception for reconciliation and SHALL NOT silently convert a failed or unreserved order to paid

#### Scenario: Recovery worker claims preparation with an unknown order outcome

- **WHEN** the timeout worker locks a due preparation before both successful outcomes are recorded
- **THEN** Checkout SHALL close preparation, obtain released or absent stock, and obtain a failed order or durable absence-fence acknowledgment before recording terminal failure

#### Scenario: Preparation success races cancellation

- **WHEN** a preparation success arrives after the attempt was closed by rejection or timeout compensation
- **THEN** Checkout SHALL continue compensation without creating payment, and delayed commands SHALL NOT leave an active stock hold or pending order after terminal failure

#### Scenario: Successful join wins after the recovery timeout

- **WHEN** both successful preparation outcomes are recorded after the timeout becomes due but before the timeout worker closes preparation
- **THEN** Checkout SHALL complete the join and schedule exactly one effective gateway creation operation without cancelling solely because the timeout passed

#### Scenario: Timeout compensation wins the workflow lock

- **WHEN** the timeout worker closes due preparation before a concurrent successful join commits
- **THEN** Checkout SHALL continue stock and order compensation and SHALL NOT schedule a gateway creation operation from the later results
