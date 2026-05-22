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

The products service MUST consume order-level `orders.created` events that include item lines and process reservation idempotently per order from within the unified catalog service runtime.

#### Scenario: Order is created

- **WHEN** the service receives `orders.created` for an order with item lines
- **THEN** it SHALL record the message idempotently and attempt reservation for each referenced product in the order

#### Scenario: Reservation is fully successful

- **WHEN** the service reserves all item lines for an order
- **THEN** it SHALL emit an order-level `inventory.reserved` event for that order

#### Scenario: Reservation cannot be completed

- **WHEN** the service cannot reserve one or more item lines for an order
- **THEN** it SHALL avoid leaving a partial active reservation for that order and emit an order-level `inventory.reservation_failed` event

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for stock-aware product reads and unified listing creation within the catalog boundary.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching product or not-found
