# Inventory Vertical Slice Plan

## Goal

Implement inventory as the source of truth for stock and reservations, separate from the product catalog.

## Scope

- Service: `services/inventory`
- Transport: gRPC
- Storage: PostgreSQL
- Responsibility: reserve, commit, and release stock safely

## Model

Use a reservation-friendly model:

- `product_id UUID PRIMARY KEY`
- `available_qty INT NOT NULL`
- `reserved_qty INT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

## Behavior

- `ReserveStock` moves quantity from `available_qty` to `reserved_qty`.
- `CommitReservation` reduces `reserved_qty` after payment success.
- `ReleaseReservation` returns reserved quantity to `available_qty` on failure or timeout.
- All state transitions should be transactional and idempotent.

## Events

- Inventory should consume `orders.item.created` events through the async backbone.
- Use inbox dedupe for repeated deliveries.
- Emit stock failure/reservation events when needed for order state transitions.
- Partition inventory consumer processing by `product_id`.
- The first real consumer flow should be `orders.item.created` -> reserve stock.

Each event should correspond to one order item and one outbox row.

## Why Reservations

- Simple `stock_quantity` is not enough for async checkout.
- Reservations prevent oversell when order creation and payment are decoupled.
- The model stays small but supports later scaling.

## Testing

- Keep tests in `services/inventory/tests/`.
- Cover reserve/commit/release behavior and validation.
- Add event-consumer tests once Kafka or a polling bridge exists.
