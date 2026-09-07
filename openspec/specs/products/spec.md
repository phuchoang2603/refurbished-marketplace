# Products

## Purpose

The products capability defines the unified catalog boundary for marketplace listings, seller ownership, stock state, and reservation behavior.

## Requirements

### Requirement: Products owns colocated stock state

The products service MUST own product catalog data together with the colocated stock state needed for marketplace listing reads and writes.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches product data for a detail or admin-oriented stock-aware catalog flow
- **THEN** the service SHALL return product data from the unified catalog boundary without requiring a separate inventory service lookup

#### Scenario: Product list is read

- **WHEN** a caller fetches a catalog product list in the first merged phase
- **THEN** the service SHALL allow that list flow to stay lighter than detail/admin surfaces and SHALL NOT require exact stock quantities everywhere

### Requirement: Products creates listings with initial stock in one logical operation

The products service MUST support authenticated seller-managed listing creation through one logical catalog write path that persists the product record together with explicit initial stock.

#### Scenario: Seller-managed listing is created

- **WHEN** a trusted internal caller creates a product for an authenticated seller-managed listing
- **THEN** the service SHALL persist the catalog fields and initial stock for that product in one logical operation

#### Scenario: Seller-managed listing is created without explicit stock

- **WHEN** a caller attempts to create a seller-managed listing without explicit initial stock
- **THEN** the service SHALL reject the request instead of silently defaulting stock

### Requirement: Product records carry seller ownership

The products service MUST persist the seller ownership identifier provided as `merchant_id` on product creation so downstream order and payment flows can attribute the listing consistently.

#### Scenario: Product is created with a merchant owner

- **WHEN** a caller creates a product with a valid `merchant_id`
- **THEN** the service SHALL store that `merchant_id` with the catalog record and return it in subsequent reads

### Requirement: Products manages reservations

The products service MUST reserve, commit, and release stock using reservation records owned inside the unified catalog boundary for each reserved order line.

#### Scenario: Stock is reserved

- **WHEN** a reservation request for an order is accepted
- **THEN** the service SHALL move quantity from available to reserved stock and persist a reservation record for the order and product

#### Scenario: Payment succeeds

- **WHEN** payment succeeds for a reservation
- **THEN** the service SHALL commit the reservation owned by that order and product

#### Scenario: Payment fails or times out

- **WHEN** payment fails or a reservation expires
- **THEN** the service SHALL release the reserved quantity back to available stock for that order-owned reservation

### Requirement: Products consumes order item events

The products service MUST consume order-level `orders.created` events that include item lines and process reservation **idempotently** per order so a Kafka delivery after the gRPC reserve command does not double-hold stock.

#### Scenario: Order is created

- **WHEN** the service receives `orders.created` for an order with item lines
- **THEN** it SHALL record the message idempotently and attempt reservation for each referenced product only when that order does not already have an active reservation from the command path

#### Scenario: Reservation is fully successful

- **WHEN** the service reserves all item lines for an order (command path or Kafka path)
- **THEN** it SHALL emit an order-level `inventory.reserved` event for that order at most once for a successful hold

#### Scenario: Reservation cannot be completed

- **WHEN** the service cannot reserve one or more item lines for an order on the Kafka path and no prior successful command-path reservation exists
- **THEN** it SHALL avoid leaving a partial active reservation for that order and emit an order-level `inventory.reservation_failed` event

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for stock-aware product reads, unified listing creation, and order-level stock reservation within the catalog boundary.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching product or not-found

#### Scenario: Reserve is requested over gRPC

- **WHEN** a documented internal caller requests reservation for an order
- **THEN** the service SHALL apply the reserve-command behavior defined for that order

### Requirement: Products supports batch lookup by IDs

The products service MUST expose a batch read that returns authoritative catalog rows for a set of product identifiers so marketplace composition (especially checkout) can re-validate many lines without N sequential single-get RPCs.

#### Scenario: Multiple known IDs are requested

- **WHEN** a caller requests products by a non-empty list of product IDs within the service’s allowed batch size
- **THEN** the service SHALL return product records for every ID that exists, from the PostgreSQL write model (including stock summary fields consistent with single-get)

#### Scenario: Some IDs are missing

- **WHEN** a batch request includes product IDs that do not exist
- **THEN** the service SHALL omit those IDs from the response rather than failing the entire batch solely because some IDs are missing

#### Scenario: Batch size exceeds the limit

- **WHEN** a caller requests more product IDs than the configured maximum batch size
- **THEN** the service SHALL reject the request as invalid

### Requirement: Products reserves stock on an internal command

The products service MUST expose an internal gRPC reservation command that holds all lines for one order idempotently so web can reserve stock before hosted payment.

#### Scenario: Reserve command succeeds

- **WHEN** a trusted caller requests reservation for an order with item lines and sufficient available stock
- **THEN** the service SHALL move quantity from available to reserved, persist reservation records for that order, and emit `inventory.reserved` once for a successful full-order hold

#### Scenario: Reserve command is retried

- **WHEN** the same order is reserved again after a successful hold
- **THEN** the service SHALL NOT increase reserved quantity again and SHALL treat the request as success for the existing reservation

#### Scenario: Reserve command cannot hold the full order

- **WHEN** one or more lines cannot be reserved
- **THEN** the service SHALL NOT leave a partial active reservation for that order, SHALL fail the command, and SHALL emit `inventory.reservation_failed` when an order-level failure signal is required for downstream consumers
