## ADDED Requirements

### Requirement: Inventory manages reservations

The inventory service MUST reserve, commit, and release stock.

#### Scenario: Stock is reserved

- **WHEN** a reservation request is accepted
- **THEN** the service SHALL move quantity from available to reserved stock

#### Scenario: Payment succeeds

- **WHEN** payment succeeds for a reservation
- **THEN** the service SHALL commit the reservation

#### Scenario: Payment fails or times out

- **WHEN** payment fails or a reservation expires
- **THEN** the service SHALL release the reserved quantity back to available stock

### Requirement: Inventory consumes order item events

The inventory service MUST consume `orders.item.created` events keyed by `product_id`.

#### Scenario: Order item is created

- **WHEN** the service receives `orders.item.created`
- **THEN** it SHALL process the message idempotently and reserve stock for that item
