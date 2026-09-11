# Payment

## Purpose

The payment capability defines how reserved orders are charged, how payment outcomes are persisted, and how order-level payment events are emitted to downstream consumers.

## Requirements

### Requirement: Payment consumes order-created events

The payment service MUST consume successful inventory reservation events and MUST NOT create a second payment transaction from that path when the transaction already exists from hosted-session create.

#### Scenario: Inventory reservation is received

- **WHEN** the service receives an inventory-reserved event for an order that already has a payment transaction
- **THEN** it SHALL deduplicate the message and SHALL NOT insert another payment transaction for that order

#### Scenario: Inventory reservation fails upstream

- **WHEN** inventory emits a reservation-failed event for an order
- **THEN** the payment service SHALL NOT create a payment transaction for that order from the failed reservation path

#### Scenario: Terminal session already recorded when reservation arrives

- **WHEN** the hosted session is already FAILED or EXPIRED and `inventory.reserved` is processed
- **THEN** the service SHALL apply the terminal payment outcome for that order if it has not already been emitted

### Requirement: Payment emits order-level outcome events after reservation

The payment service MUST emit order-level payment success and failure events from hosted gateway outcomes only for orders whose inventory reservation and payment state allow a terminal outcome to be published.

#### Scenario: Reserved order payment completes

- **WHEN** an order payment transaction for a reserved order succeeds or fails through the hosted gateway flow
- **THEN** the service SHALL write a corresponding order-level outbox event for downstream consumers

### Requirement: Payment persists inbox and outbox state

The payment service MUST store inbox and outbox records in PostgreSQL.

#### Scenario: Message is processed

- **WHEN** the service processes a Kafka message
- **THEN** it SHALL record the message in the inbox before advancing offsets

### Requirement: Payment creates hosted payment sessions by order identifier

The payment service MUST create a hosted payment session using `order_id` as the idempotency anchor and MUST return hosted-session metadata that the web edge can use to redirect the buyer.

#### Scenario: Hosted payment session is requested for a new order

- **WHEN** the web edge requests a hosted payment session for an order with buyer, merchant, amount, optional shipping, and return context
- **THEN** the payment service SHALL persist the hosted session and payment transaction and return session metadata including `order_id`, `payment_session_id`, and the return URL

#### Scenario: Hosted payment session is requested again for the same order

- **WHEN** the web edge repeats the hosted payment session request for an order that already has a stored PENDING session
- **THEN** the payment service SHALL return stored session metadata for that `order_id` instead of creating a second independent payment identity

### Requirement: Payment accepts hosted gateway outcome callbacks

The payment service MUST accept hosted gateway payment outcomes over its internal gRPC contract and update payment state idempotently.

#### Scenario: Gateway reports a terminal payment result

- **WHEN** the web edge forwards a successful or failed terminal payment result for an order or payment session
- **THEN** the payment service SHALL update the corresponding payment state and emit the order-level payment outcome expected by downstream consumers

#### Scenario: Gateway repeats a terminal payment result

- **WHEN** the web edge forwards the same terminal callback again
- **THEN** the payment service SHALL treat the repeat as idempotent and SHALL NOT emit duplicate terminal outcomes for downstream consumers

### Requirement: Payment expires abandoned hosted sessions

The payment service MUST periodically expire PENDING hosted payment sessions whose `expires_at` is in the past, mark them EXPIRED (distinct from FAILED), and emit `payment.failed` so downstream services can release reserved stock and fail the order.

#### Scenario: Pending session past expires_at is swept

- **WHEN** a hosted payment session remains PENDING after its `expires_at` timestamp
- **THEN** the payment service SHALL mark the session EXPIRED and emit an order-level `payment.failed` outbox event

#### Scenario: Gateway reports expired as distinct from declined

- **WHEN** the hosted gateway posts an EXPIRED outcome for a PENDING session
- **THEN** the payment service SHALL store status EXPIRED (not FAILED) and SHALL emit `payment.failed` for downstream consumers

### Requirement: Payment snapshots commerce facts when creating a hosted session

The payment service MUST persist buyer, merchant, amount, currency, optional shipping address, and optional line-item snapshot on hosted-session create so a later fraud gateway can score the attempt without waiting for Kafka.

#### Scenario: Session create includes charge facts

- **WHEN** the web edge requests a hosted payment session with `order_id`, buyer, merchant, total cents, currency, and return URL
- **THEN** the payment service SHALL store those facts with the session and SHALL create the order payment transaction in the same operation

#### Scenario: Session create includes shipping when provided

- **WHEN** the web edge supplies a shipping address on hosted-session create
- **THEN** the payment service SHALL persist that shipping address on the session

### Requirement: Hosted payment session is one-shot per order

The payment service MUST treat `order_id` as a single payment attempt. It MUST NOT mint a new `payment_session_id` after the session is FAILED, EXPIRED, or SUCCEEDED.

#### Scenario: Repeat create while pending

- **WHEN** the web edge repeats hosted-session create for an order whose session is still PENDING
- **THEN** the service SHALL return the existing session metadata without creating a second payment identity

#### Scenario: Repeat create after terminal session

- **WHEN** the web edge requests hosted-session create for an order whose session is SUCCEEDED, FAILED, or EXPIRED
- **THEN** the service SHALL reject the request and SHALL NOT refresh or replace the session
