## ADDED Requirements

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

## MODIFIED Requirements

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
