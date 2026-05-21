## MODIFIED Requirements

### Requirement: Inventory manages reservations

The inventory behavior MUST reserve, commit, and release stock using reservation records owned inside the unified catalog service for each reserved order line.

#### Scenario: Stock is reserved

- **WHEN** a reservation request for an order is accepted
- **THEN** the service SHALL move quantity from available to reserved stock and persist a reservation record for the order and product

#### Scenario: Payment succeeds

- **WHEN** payment succeeds for a reservation
- **THEN** the service SHALL commit the reservation owned by that order and product

#### Scenario: Payment fails or times out

- **WHEN** payment fails or a reservation expires
- **THEN** the service SHALL release the reserved quantity back to available stock for that order-owned reservation

### Requirement: Inventory consumes order item events

The inventory behavior MUST consume order-level `orders.created` events that include item lines and process reservation idempotently per order from within the unified catalog service runtime.

#### Scenario: Order is created

- **WHEN** the service receives `orders.created` for an order with item lines
- **THEN** it SHALL record the message idempotently and attempt reservation for each referenced product in the order

#### Scenario: Reservation is fully successful

- **WHEN** the service reserves all item lines for an order
- **THEN** it SHALL emit an order-level `inventory.reserved` event for that order

#### Scenario: Reservation cannot be completed

- **WHEN** the service cannot reserve one or more item lines for an order
- **THEN** it SHALL avoid leaving a partial active reservation for that order and emit an order-level `inventory.reservation_failed` event

## ADDED Requirements

### Requirement: Inventory state is initialized during listing creation

The inventory behavior MUST initialize stock records during unified product listing creation instead of waiting for a separate inventory bootstrap flow, and that initialization MUST use the explicit initial stock supplied by the caller.

#### Scenario: Listing is created with initial stock

- **WHEN** a product listing is created through the unified catalog boundary
- **THEN** the service SHALL initialize the matching stock record as part of that listing creation flow
