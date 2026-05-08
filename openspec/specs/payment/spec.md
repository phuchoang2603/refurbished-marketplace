## ADDED Requirements

### Requirement: Payment consumes order item events

The payment service MUST consume `orders.item.created` events and create payment transactions.

#### Scenario: Order item is received

- **WHEN** the service receives `orders.item.created`
- **THEN** it SHALL deduplicate the message and create or update the item transaction

### Requirement: Payment emits item outcome events

The payment service MUST emit `payment.item.succeeded` and `payment.item.failed` events through its outbox path.

#### Scenario: Payment completes

- **WHEN** a payment transaction succeeds or fails
- **THEN** the service SHALL write a corresponding outbox event for downstream consumers

### Requirement: Payment persists inbox and outbox state

The payment service MUST store inbox and outbox records in PostgreSQL.

#### Scenario: Message is processed

- **WHEN** the service processes a Kafka message
- **THEN** it SHALL record the message in the inbox before advancing offsets
