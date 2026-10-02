## MODIFIED Requirements

### Requirement: Payment emits order-level outcome events after reservation

Payment MUST create hosted sessions only after Checkout reports that stock is reserved and MUST publish definitive, correlated payment outcomes to Checkout. Payment MUST NOT independently direct Orders or Inventory to finalize a checkout.

#### Scenario: Reserved order payment completes

- **WHEN** a hosted session for a confirmed reservation succeeds or fails
- **THEN** Payment SHALL persist the result and a correlated outcome for Checkout in one transaction

#### Scenario: Session command arrives without a reservation authorization

- **WHEN** a session-create command lacks Checkout's valid reservation context
- **THEN** Payment SHALL reject the command without creating a payable session

### Requirement: Payment persists inbox and outbox state

The payment service MUST store inbox and outbox records in PostgreSQL and MUST atomically persist processed command identities, session/outcome changes, and emitted results.

#### Scenario: Message is processed

- **WHEN** the service processes a checkout command
- **THEN** it SHALL commit its inbox record, the corresponding payment state, and any result outbox row before advancing offsets

### Requirement: Payment creates hosted payment sessions by order identifier

Payment MUST create a hosted payment session using `order_id` as the single-attempt idempotency anchor and MUST publish correlated session metadata for Checkout to expose to the buyer only after reservation.

#### Scenario: Hosted payment session is requested for a new order

- **WHEN** Checkout requests a hosted session with buyer, merchant, amount, shipping, line items, return context, and confirmed reservation
- **THEN** Payment SHALL persist one hosted session and transaction and publish a correlated ready result including order id, session id, and return URL

#### Scenario: Hosted payment session is requested again for the same order

- **WHEN** Checkout repeats a request for an order with a stored PENDING session and identical facts
- **THEN** Payment SHALL publish or return the existing session metadata without minting a new payment identity

#### Scenario: Checkout facts conflict

- **WHEN** a repeat request changes an order's charge or party facts
- **THEN** Payment SHALL reject it without changing the existing session or transaction

### Requirement: Payment accepts hosted gateway outcome callbacks

Payment MUST authenticate the gateway callback as supported by the configured gateway integration, match the session identity, store verified outcomes idempotently, and notify Checkout. A late or contradictory success MUST be retained as a financial exception, never silently converted to failure or automatically mark an unreserved order paid.

#### Scenario: Gateway reports a terminal payment result

- **WHEN** web forwards a verified success or failure callback for the matching session
- **THEN** Payment SHALL persist one definitive result and publish a correlated outcome to Checkout

#### Scenario: Gateway repeats a terminal payment result

- **WHEN** the gateway repeats the same terminal callback
- **THEN** Payment SHALL acknowledge the repeat without emitting an additional effective outcome

#### Scenario: Gateway reports contradictory late success

- **WHEN** a verified success arrives after the session was cancelled or expired
- **THEN** Payment SHALL preserve evidence of the success and report a reconciliation exception to Checkout rather than discard it

### Requirement: Payment expires abandoned hosted sessions

Payment MUST expire abandoned PENDING sessions after their deadline and distinguish EXPIRED from FAILED. It MUST report expiry to Checkout for compensation only when payment is definitively not captured; uncertain gateway outcomes MUST be reported as unresolved instead of as failure.

#### Scenario: Pending session past expires_at is swept

- **WHEN** Payment confirms a PENDING session has expired without capture
- **THEN** it SHALL mark the session EXPIRED and publish a correlated definitive expiry result to Checkout

#### Scenario: Gateway reports expired as distinct from declined

- **WHEN** the deadline passes but the gateway outcome is uncertain
- **THEN** Payment SHALL publish or expose an unresolved status and SHALL NOT assert that no payment occurred

### Requirement: Payment snapshots commerce facts when creating a hosted session

Payment MUST persist nested buyer and merchant snapshots (marketplace ids, optional buyer email), amount, currency, shipping address, and line items (product id, name, quantity, unit price) supplied by Checkout with the hosted session, without calling Users to hydrate party fields.

#### Scenario: Session create includes charge and party facts

- **WHEN** Checkout requests a session with order id, buyer and merchant ids, total cents, currency, usable shipping, named line items, and return URL
- **THEN** Payment SHALL store the facts and create the corresponding transaction in the same operation

#### Scenario: Session create omits required commerce facts

- **WHEN** Checkout requests a session without a usable shipping address (line1, city, postal code, and country) or buyer or merchant id
- **THEN** Payment SHALL reject it and SHALL NOT create a payable session

### Requirement: Hosted payment session is one-shot per order

Payment MUST treat `order_id` as one payment attempt and MUST NOT mint another `payment_session_id` after SUCCEEDED, FAILED, or EXPIRED.

#### Scenario: Repeat create while pending

- **WHEN** Checkout repeats a hosted-session command while the same order's session remains PENDING
- **THEN** Payment SHALL return or republish the existing session identity without creating another

#### Scenario: Repeat create after terminal session

- **WHEN** Checkout requests a new hosted session for an order with a terminal attempt
- **THEN** Payment SHALL reject the command and SHALL NOT refresh or replace the session

## REMOVED Requirements

### Requirement: Payment consumes order-created events

**Reason**: `inventory.reserved` no longer acts as a catch-up/inbox dependency; Checkout requests a session only after observing a committed reservation result.

**Migration**: Remove the `inventory.reserved` payment consumer as part of the fresh start and use idempotent Checkout session commands and result events.
