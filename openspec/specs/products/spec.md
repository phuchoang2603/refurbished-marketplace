## ADDED Requirements

### Requirement: Products owns colocated stock state

The products service MUST own product catalog data together with the colocated stock state needed for marketplace listing reads and writes.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches product data for a detail or admin-oriented stock-aware catalog flow
- **THEN** the service SHALL return product data from the unified catalog boundary without requiring a separate inventory service lookup

#### Scenario: Product list is read

- **WHEN** a caller fetches a catalog product list in the first merged phase
- **THEN** the service SHALL allow that list flow to stay lighter than detail/admin surfaces and SHALL NOT require exact stock quantities everywhere

### Requirement: Products creates listings with initial stock in one logical operation

The products service MUST support creating a product and initializing its stock state within one logical catalog write path, and it MUST require initial stock explicitly in that create operation.

#### Scenario: Listing is created

- **WHEN** a caller creates a new product listing with an initial stock quantity
- **THEN** the service SHALL persist the product record and its initial stock state without requiring a second downstream service call

#### Scenario: Listing is created without initial stock

- **WHEN** a caller attempts to create a new product listing without an explicit initial stock quantity
- **THEN** the service SHALL reject the request instead of silently defaulting stock

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

## MODIFIED Requirements

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for stock-aware product reads and unified listing creation within the catalog boundary.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching product or not-found
