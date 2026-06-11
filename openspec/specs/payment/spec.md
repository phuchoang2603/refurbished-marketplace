# Payment

## Purpose

The payment capability defines how reserved orders are charged, how payment outcomes are persisted, and how order-level payment events are emitted to downstream consumers.

## Requirements

### Requirement: Payment consumes order-created events

The payment service MUST consume successful inventory reservation events and create one payment transaction per order.

#### Scenario: Inventory reservation is received

- **WHEN** the service receives an inventory-reserved event for an order
- **THEN** it SHALL deduplicate the message and create or update the order payment transaction

#### Scenario: Inventory reservation fails upstream

- **WHEN** inventory emits a reservation-failed event for an order
- **THEN** the payment service SHALL NOT create a payment transaction for that order from the failed reservation path

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

The payment service MUST create or reuse a hosted payment session using `order_id` as the idempotency anchor and MUST return hosted-session metadata that the web edge can use to redirect the buyer.

#### Scenario: Hosted payment session is requested for a new order

- **WHEN** the web edge requests a hosted payment session for an order with buyer, optional shipping, and return context
- **THEN** the payment service SHALL persist the hosted session state and return session metadata including `order_id`, `payment_session_id`, and return or cancel URLs

#### Scenario: Hosted payment session is requested again for the same order

- **WHEN** the web edge repeats the hosted payment session request for an order that already has a stored session
- **THEN** the payment service SHALL return the same stored session metadata instead of creating a duplicate

### Requirement: Payment accepts hosted gateway outcome callbacks

The payment service MUST accept hosted gateway payment outcomes over its internal gRPC contract and update payment state idempotently.

#### Scenario: Gateway reports a terminal payment result

- **WHEN** the web edge forwards a successful or failed terminal payment result for an order or payment session
- **THEN** the payment service SHALL update the corresponding payment state and emit the order-level payment outcome expected by downstream consumers

#### Scenario: Gateway repeats a terminal payment result

- **WHEN** the web edge forwards the same terminal callback again
- **THEN** the payment service SHALL treat the repeat as idempotent and SHALL NOT emit duplicate terminal outcomes for downstream consumers
