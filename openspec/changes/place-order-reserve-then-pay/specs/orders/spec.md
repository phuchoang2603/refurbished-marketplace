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

### Requirement: Place order reserves stock before success

The orders capability MUST reserve all order lines through the products catalog **before** treating place-order as successful so the caller does not receive an `order_id` for payment until stock is held.

#### Scenario: Stock is available

- **WHEN** place-order persists a new order and products accepts the reservation for every line
- **THEN** the service SHALL return the order as placed and SHALL leave the order in a pending unpaid state until a payment outcome arrives

#### Scenario: Stock is insufficient or reservation fails

- **WHEN** products cannot reserve the full order
- **THEN** the service SHALL NOT return a successful place-order for payment, SHALL NOT leave a partial active reservation for that order, and SHALL fail or cancel the order record created for that intent

#### Scenario: Retry after successful reserve

- **WHEN** place-order is retried with the same buyer and `idempotency_key` after stock was already reserved
- **THEN** the service SHALL return the existing reserved order without reserving additional quantity

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

- **WHEN** place-order has reserved stock over gRPC
- **THEN** the caller SHALL be able to proceed to hosted payment without waiting for `orders.created` to be consumed
