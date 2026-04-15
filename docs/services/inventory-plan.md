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

## Optional future: inventory-first payment gate (`orders.pay_ready`)

If you want to enforce **reserve-then-pay**, inventory can also act as a gate:

- Ingress stays `orders.item.created` keyed by `product_id`.
- Inventory reserves each line idempotently and tracks completion for the parent `order_id`.
- Once all lines for the order are reserved, inventory emits exactly one `orders.pay_ready` event keyed by `order_id` for payment to consume.

Implementation notes (high level):

- **Completion metadata:** include `lines_total` on each `orders.item.created` payload so inventory can know when the set is complete without calling `orders`.
- **Durable state:** use two tables in the inventory DB:
  - `inventory_line_reservations` keyed by `(order_id, order_item_id)` for idempotency and audit
  - `inventory_order_gate` keyed by `order_id` to serialize completion checks (`SELECT … FOR UPDATE`) and ensure only one gate event is emitted
- **Outbox for gate:** `inventory_outbox` mirrors `orders_outbox`, but for `orders.pay_ready` rows set `aggregate_id = order_id` so Kafka key is `order_id`.

## Why Reservations

- Simple `stock_quantity` is not enough for async checkout.
- Reservations prevent oversell when order creation and payment are decoupled.
- The model stays small but supports later scaling.

## Testing

- Keep tests in `services/inventory/tests/`.
- Cover reserve/commit/release behavior and validation.
- Add event-consumer tests once Kafka or a polling bridge exists.
