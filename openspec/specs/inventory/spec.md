# Inventory

## Purpose

Own marketplace available and reserved quantity, order-level reservations, and reservation Kafka as a dedicated inventory service with its own Postgres, not colocated in the products runtime.

## Requirements

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

### Requirement: Inventory owns the stock ledger

The inventory service SHALL persist `available_qty` and `reserved_qty` per product identifier. It SHALL NOT store listing documents (name, description, price, merchant). Product identifiers SHALL NOT require a foreign key to a catalog SQL table.

#### Scenario: Stock row exists without a SQL catalog table

- **WHEN** ProductCreated is consumed with a product id that exists only as a catalog document elsewhere
- **THEN** inventory persists a stock row keyed by that id

#### Scenario: Unknown identity cannot mutate stock

- **WHEN** a caller other than documented internal identities invokes mutating inventory RPCs
- **THEN** the request is rejected at the mesh or service boundary

### Requirement: Inventory seeds stock from ProductCreated

Inventory SHALL consume ProductCreated on `products.created` in a consumer group independent of the future search projector and of inventory's reservation/payment consumer. It SHALL validate an explicit non-negative initial quantity and atomically persist the event id in its inbox with a stock row whose reserved quantity is zero. EnsureStock SHALL remain an internal operation and SHALL NOT be exposed over gRPC.

#### Scenario: First seed succeeds

- **WHEN** inventory consumes a valid ProductCreated for an unseeded product
- **THEN** it SHALL commit the inbox record and initial stock together before acknowledging processing

#### Scenario: Duplicate seed for the same product

- **WHEN** a creation event is redelivered or replayed after the product has been seeded, including after reservations have changed its quantities
- **THEN** inventory SHALL NOT add, reset, or overwrite stock and SHALL treat a matching seed intent as successfully processed

#### Scenario: Conflicting seed intent

- **WHEN** a creation event attempts to seed an existing product with conflicting creation intent
- **THEN** inventory SHALL preserve the existing ledger and surface a processing error rather than reset quantity

#### Scenario: Missing or invalid initial quantity

- **WHEN** ProductCreated omits initial quantity or carries a negative quantity
- **THEN** inventory SHALL reject processing without seeding stock or recording successful consumption

#### Scenario: Consumer fails before commit

- **WHEN** creation-event processing fails before its database transaction commits
- **THEN** neither successful inbox consumption nor initial stock SHALL be committed, and transient failures SHALL remain retryable without deleting the catalog listing

#### Scenario: Consumer retries after commit

- **WHEN** the stock transaction committed but delivery is retried before acknowledgement completed
- **THEN** inventory SHALL recognize the committed event and SHALL NOT seed stock again

#### Scenario: Creation consumption is isolated from reservation topics

- **WHEN** ProductCreated processing returns a retryable or validation error
- **THEN** inventory's `orders.created` and payment-outcome consumer SHALL continue independently and SHALL NOT stall reservation settlement on that error

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

### Requirement: Inventory exposes stock reads

Inventory SHALL return available and reserved quantities for one product id and for a batch of ids so web can show PDP qty until a projected read model exists. Products SHALL NOT use these RPCs.

#### Scenario: Get stock by id

- **WHEN** web requests stock for an existing product id
- **THEN** inventory returns available and reserved quantities

#### Scenario: Missing stock row

- **WHEN** a caller requests stock for an id with no row
- **THEN** inventory returns not-found (callers SHALL NOT invent quantity)

#### Scenario: Batch omits missing ids

- **WHEN** a batch stock request includes ids with no row
- **THEN** those ids are omitted rather than failing the whole batch
