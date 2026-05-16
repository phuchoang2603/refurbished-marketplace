## ADDED Requirements

### Requirement: Orders are merchant-scoped records

The orders capability MUST persist each created order with exactly one caller-supplied `merchant_id` and MUST accept only merchant-scoped order creation requests.

#### Scenario: Single-merchant order is created

- **WHEN** a caller creates an order with a merchant identifier and items that all belong to that merchant grouping
- **THEN** the orders service SHALL persist one order record for that merchant and the corresponding order items

#### Scenario: Merchant identifier is omitted from order creation

- **WHEN** a caller submits an order creation request without a merchant identifier
- **THEN** the orders service SHALL reject the request as invalid

#### Scenario: Mixed-merchant checkout is handled upstream

- **WHEN** a checkout flow contains items from more than one merchant
- **THEN** the upstream caller SHALL split that work into separate merchant-scoped order creation requests before calling the orders service

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
