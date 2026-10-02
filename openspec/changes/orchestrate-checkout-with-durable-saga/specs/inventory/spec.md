## MODIFIED Requirements

### Requirement: Inventory manages reservations

Inventory MUST reserve, commit, and release stock using order-owned reservation records, with idempotent, correlated results for Checkout. A release/cancel decision MUST fence later reserve commands for that order so a delayed command cannot recreate a hold after checkout failure.

#### Scenario: Stock is reserved

- **WHEN** Checkout requests reservation for an order with available stock
- **THEN** Inventory SHALL move quantity from available to reserved, persist order-owned records, and report a correlated reserved result

#### Scenario: Payment succeeds

- **WHEN** Checkout requests stock commit after confirmed payment success
- **THEN** Inventory SHALL commit that order's active reservations once and report stock committed

#### Scenario: Payment fails or times out

- **WHEN** Checkout requests release after a definitive failed or cancelled payment
- **THEN** Inventory SHALL release that order's active reservations once and report stock released or absent

#### Scenario: Cancel precedes delayed reserve

- **WHEN** Inventory durably accepts a cancellation before an earlier reserve command is delivered
- **THEN** Inventory SHALL reject the later reserve without holding stock and SHALL report the reservation as absent or released

## ADDED Requirements

### Requirement: Inventory processes Checkout reservation commands

Inventory MUST consume Checkout's correlated reserve, commit, and release commands idempotently, and report each result from its own committed transaction. An `orders.created` event MUST NOT independently reserve stock for a saga checkout.

#### Scenario: Order-created event arrives without Checkout command

- **WHEN** Inventory observes an `orders.created` event for a saga checkout before a reserve command
- **THEN** it SHALL NOT reserve stock or emit `inventory.reserved` from that event alone

#### Scenario: Reservation command is redelivered

- **WHEN** Checkout's reserve command is redelivered for the same order
- **THEN** Inventory SHALL NOT double-hold stock and SHALL report the existing definitive reservation result

#### Scenario: Reservation cannot be completed

- **WHEN** one or more lines cannot be fully reserved
- **THEN** Inventory SHALL leave no partial active hold and SHALL report a correlated reservation failure

#### Scenario: Web attempts a direct checkout reserve

- **WHEN** web attempts to reserve stock for checkout using the removed synchronous path
- **THEN** the service or mesh SHALL reject the mutation and SHALL NOT create a hold

## REMOVED Requirements

### Requirement: Inventory consumes order item events

**Reason**: `orders.created` must not race the Checkout coordinator and create reservations that Checkout has not requested.

**Migration**: Remove the checkout auto-reserve consumer when replacing the old flow, and use correlated Checkout reserve commands on the fresh dataset.

### Requirement: Inventory exposes ReserveStock over gRPC

**Reason**: Web's direct reserve call bypasses the durable Checkout workflow; authorized checkout mutation is now an idempotent command.

**Migration**: Keep stock-read RPCs, revoke web's checkout reserve permission, and use Checkout reserve/settle commands.
