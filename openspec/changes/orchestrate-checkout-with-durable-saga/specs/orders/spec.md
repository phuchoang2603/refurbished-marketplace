## MODIFIED Requirements

### Requirement: Orders write order-level outbox events

The orders capability MUST write one outbox row per created order, not one outbox row per order item, and the order-created payload MUST include the item lines required for downstream consumers. Checkout order creation MUST be driven by Checkout's idempotent command, and checkout reservation MUST wait for Checkout to observe the correlated order-created result.

#### Scenario: Merchant order is created

- **WHEN** an order is persisted successfully in response to Checkout's command
- **THEN** the service SHALL store the order, its items, and one correlated order-created outbox row in the same transaction

#### Scenario: Order-created payload is published

- **WHEN** the service writes the order-created outbox row
- **THEN** the payload SHALL include the checkout and order identifiers, buyer identifier, merchant identifier, total amount, and each order item's product identifier and quantity

#### Scenario: Place-order returns before Kafka consume

- **WHEN** Orders has persisted a new checkout order but its order-created event is delayed
- **THEN** Checkout SHALL remain pending and SHALL NOT reserve stock before observing the correlated order-created result

### Requirement: Place order persists without calling products

Orders MUST persist merchant-scoped orders and emit `orders.created` without calling Products or Inventory. Checkout MUST request reservation only after the correlated order-created result is committed.

#### Scenario: Stock is available

- **WHEN** Orders processes a checkout order command
- **THEN** Orders SHALL persist the pending order and report its creation without synchronously reserving stock or creating payment

#### Scenario: Retry after successful persist

- **WHEN** the same checkout order command is retried after the order was created
- **THEN** Orders SHALL return or re-emit the original correlated outcome without inserting another order

## ADDED Requirements

### Requirement: Orders finalize checkout only on saga commands

Orders MUST finalize checkout orders only on correlated commands from Checkout after stock settlement. Orders MUST NOT independently finalize checkout orders from inventory reservation or payment outcome events.

#### Scenario: Stock was committed after payment

- **WHEN** Checkout commands paid finalization after Inventory confirms stock commit
- **THEN** Orders SHALL update the pending order to paid and emit a correlated finalization result

#### Scenario: Stock was released after definitive failure

- **WHEN** Checkout commands failed finalization after Inventory confirms stock release or absence
- **THEN** Orders SHALL update the pending order to failed and emit a correlated finalization result

#### Scenario: Conflicting finalization command arrives

- **WHEN** Orders receives a duplicate or contradictory finalization for an already terminal order
- **THEN** Orders SHALL acknowledge exact duplicates without mutation and reject conflicting outcomes for reconciliation

## REMOVED Requirements

### Requirement: Orders consume payment results

**Reason**: Independently applying payment and reservation outcome events can mark the order terminal before stock settlement and competes with the Checkout saga.

**Migration**: On the fresh dataset, replace direct result consumption with Checkout's paid/failed finalization commands and correlated acknowledgements; no legacy order migration is required.
