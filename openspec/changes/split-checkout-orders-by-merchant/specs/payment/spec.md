## ADDED Requirements

### Requirement: Payment consumes order-created events

The payment service MUST consume order-created events and create one payment transaction per order.

#### Scenario: Order is received

- **WHEN** the service receives an order-created event
- **THEN** it SHALL deduplicate the message and create or update the order payment transaction

### Requirement: Payment emits order-level outcome events

The payment service MUST emit order-level payment success and failure events through its outbox path.

#### Scenario: Payment completes

- **WHEN** an order payment transaction succeeds or fails
- **THEN** the service SHALL write a corresponding order-level outbox event for downstream consumers

## MODIFIED Requirements

### Requirement: Payment persists inbox and outbox state

The payment service MUST store inbox and outbox records in PostgreSQL.

#### Scenario: Message is processed

- **WHEN** the service processes a Kafka message
- **THEN** it SHALL record the message in the inbox before advancing offsets

## REMOVED Requirements

### Requirement: Payment consumes order item events

**Reason**: Payment is no longer created per order item once checkout fans out into one merchant-scoped order at a time.
**Migration**: Consume the new order-created event and key payment idempotency by `order_id`.

### Requirement: Payment emits item outcome events

**Reason**: Downstream order updates now happen once per order instead of once per order item.
**Migration**: Emit order-level payment outcome events keyed by `order_id` and update consumers accordingly.
