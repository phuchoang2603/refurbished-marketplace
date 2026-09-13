## Purpose

Own marketplace available and reserved quantity, order-level reservations, and reservation Kafka as a dedicated inventory service with its own Postgres, not colocated in the products runtime.

## ADDED Requirements

### Requirement: Inventory owns the stock ledger

The inventory service SHALL persist `available_qty` and `reserved_qty` per product identifier. It SHALL NOT store listing documents (name, description, price, merchant). Product identifiers SHALL NOT require a foreign key to a catalog SQL table.

#### Scenario: Stock row exists without a SQL catalog table

- **WHEN** ProductCreated is consumed with a product id that exists only as a catalog document elsewhere
- **THEN** inventory persists a stock row keyed by that id

#### Scenario: Unknown identity cannot mutate stock

- **WHEN** a caller other than documented internal identities invokes mutating inventory RPCs
- **THEN** the request is rejected at the mesh or service boundary

### Requirement: Inventory seeds stock from ProductCreated

Inventory SHALL consume ProductCreated on `products.created` in a consumer group independent of the future search projector. It SHALL validate an explicit non-negative initial quantity and atomically persist the event id in its inbox with a stock row whose reserved quantity is zero. EnsureStock SHALL remain an internal operation and SHALL NOT be exposed over gRPC.

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

### Requirement: Inventory manages reservations

The inventory service MUST reserve, commit, and release stock using reservation records it owns for each reserved order line.

#### Scenario: Stock is reserved

- **WHEN** a reservation request for an order is accepted
- **THEN** the service SHALL move quantity from available to reserved and persist a reservation record for the order and product

#### Scenario: Payment succeeds

- **WHEN** payment succeeds for a reservation
- **THEN** the service SHALL commit the reservation owned by that order and product

#### Scenario: Payment fails or times out

- **WHEN** payment fails or a reservation expires
- **THEN** the service SHALL release the reserved quantity back to available stock for that order-owned reservation

### Requirement: Inventory consumes order item events

The inventory service MUST consume order-level `orders.created` events that include item lines and process reservation **idempotently** per order so a Kafka delivery after the gRPC reserve command does not double-hold stock.

#### Scenario: Order is created

- **WHEN** the service receives `orders.created` for an order with item lines
- **THEN** it SHALL record the message idempotently and attempt reservation for each referenced product only when that order does not already have an active reservation from the command path

#### Scenario: Reservation is fully successful

- **WHEN** the service reserves all item lines for an order (command path or Kafka path)
- **THEN** it SHALL emit an order-level `inventory.reserved` event for that order at most once for a successful hold

#### Scenario: Reservation cannot be completed

- **WHEN** the service cannot reserve one or more item lines for an order on the Kafka path and no prior successful command-path reservation exists
- **THEN** it SHALL avoid leaving a partial active reservation for that order and emit an order-level `inventory.reservation_failed` event

### Requirement: Inventory exposes ReserveStock over gRPC

The inventory service MUST expose an internal gRPC reservation command that holds all lines for one order idempotently so web can reserve stock before hosted payment.

#### Scenario: Reserve command succeeds

- **WHEN** a trusted caller requests reservation for an order with item lines and sufficient available stock
- **THEN** the service SHALL move quantity from available to reserved, persist reservation records for that order, and emit `inventory.reserved` once for a successful full-order hold

#### Scenario: Reserve command is retried

- **WHEN** the same order is reserved again after a successful hold
- **THEN** the service SHALL NOT increase reserved quantity again and SHALL treat the request as success for the existing reservation

#### Scenario: Reserve command cannot hold the full order

- **WHEN** one or more lines cannot be reserved
- **THEN** the service SHALL NOT leave a partial active reservation for that order, SHALL fail the command, and SHALL emit `inventory.reservation_failed` when an order-level failure signal is required for downstream consumers

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
