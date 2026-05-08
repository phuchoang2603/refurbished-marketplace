## ADDED Requirements

### Requirement: Orders write line-item outbox events

The orders service MUST write one outbox row per order item when an order is created.

#### Scenario: Order is created

- **WHEN** a checkout order is persisted
- **THEN** the service SHALL store the order, line items, and outbox rows in the same transaction

### Requirement: Orders consume payment results

The orders service MUST consume payment success and failure events and update order state.

#### Scenario: Payment succeeds

- **WHEN** the service receives `payment.item.succeeded`
- **THEN** the order SHALL be updated to the paid state

#### Scenario: Payment fails

- **WHEN** the service receives `payment.item.failed`
- **THEN** the order SHALL be updated to the failed state
