## ADDED Requirements

### Requirement: Place order is idempotent per buyer intent

The orders capability MUST accept a caller-supplied `idempotency_key` on place-order and MUST persist it with a uniqueness constraint per buyer so retries return the original order instead of inserting a second row.

#### Scenario: First place-order for an intent

- **WHEN** a caller places an order with a new `idempotency_key` for that buyer
- **THEN** the service SHALL persist one order and associated items for that key

#### Scenario: Retry with the same buyer and key

- **WHEN** the same buyer places an order again with the same `idempotency_key`
- **THEN** the service SHALL return the existing order and SHALL NOT insert a second order row

#### Scenario: Same key with a conflicting body

- **WHEN** the same buyer reuses an `idempotency_key` with a different merchant or different item lines
- **THEN** the service SHALL reject the request as a conflict without creating another order

#### Scenario: Idempotency key is omitted

- **WHEN** a caller places an order without an `idempotency_key`
- **THEN** the service SHALL reject the request as invalid

### Requirement: Place order persists without calling products

The orders capability MUST persist merchant-scoped orders and emit `orders.created`. It MUST NOT call products to reserve stock.

#### Scenario: Stock is available

- **WHEN** place-order persists a new order
- **THEN** the service SHALL return the pending order so the caller can reserve stock and open hosted payment

#### Scenario: Retry after successful persist

- **WHEN** place-order is retried with the same buyer and `idempotency_key` after the order was created
- **THEN** the service SHALL return the existing order without inserting another row

## MODIFIED Requirements

### Requirement: Orders write order-level outbox events

The orders capability MUST write one outbox row per created order, not one outbox row per order item, and the order-created payload MUST include the item lines required for downstream consumers. Reservation for checkout MUST NOT depend on that outbox row being consumed before place-order returns.

#### Scenario: Merchant order is created

- **WHEN** an order is persisted successfully
- **THEN** the service SHALL store the order, its items, and one order-created outbox row in the same transaction

#### Scenario: Order-created payload is published

- **WHEN** the service writes the order-created outbox row
- **THEN** the payload SHALL include the order identifier, buyer identifier, merchant identifier, total amount, and each order item's product identifier and quantity

#### Scenario: Place-order returns before Kafka consume

- **WHEN** place-order has persisted the order
- **THEN** the caller SHALL be able to reserve stock over gRPC and proceed to hosted payment without waiting for `orders.created` to be consumed
