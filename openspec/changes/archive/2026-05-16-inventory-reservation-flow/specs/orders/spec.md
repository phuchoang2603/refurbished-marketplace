## MODIFIED Requirements

### Requirement: Orders write order-level outbox events

The orders capability MUST write one outbox row per created order, not one outbox row per order item, and the order-created payload MUST include the item lines required for downstream reservation.

#### Scenario: Merchant order is created

- **WHEN** an order is persisted successfully
- **THEN** the service SHALL store the order, its items, and one order-created outbox row in the same transaction

#### Scenario: Order-created payload is published

- **WHEN** the service writes the order-created outbox row
- **THEN** the payload SHALL include the order identifier, buyer identifier, merchant identifier, total amount, and each order item's product identifier and quantity

### Requirement: Orders consume payment results

The orders service MUST consume order-level payment success and failure events and inventory reservation failure events and update order state.

#### Scenario: Payment succeeds

- **WHEN** the service receives the order-level payment success event for an order
- **THEN** the order SHALL be updated to the paid state

#### Scenario: Payment fails

- **WHEN** the service receives the order-level payment failure event for an order
- **THEN** the order SHALL be updated to the failed state

#### Scenario: Inventory reservation fails

- **WHEN** the service receives the order-level inventory reservation failure event for an order
- **THEN** the order SHALL be updated to the failed state
