# Orders Vertical Slice Plan

## Goal

Implement checkout order headers and per-item outbox events that drive inventory and payment asynchronously.

## Scope

- Service: `services/orders`
- Transport: gRPC
- Storage: PostgreSQL (`orders_db`)
- Events: `orders.item.created`

## Data Model

Keep orders as headers plus line items:

- `orders.id`
- `orders.buyer_user_id`
- `orders.status`
- `orders.total_cents`
- `order_items.product_id`
- `order_items.merchant_id`
- `order_items.quantity`
- `order_items.unit_price_cents`
- `order_items.line_total_cents`

## Eventing

- Write one outbox row per order item in the same transaction as the order write.
- Use `product_id` as the Kafka partition key so inventory sees stable per-product ordering.
- Snapshot `merchant_id` into the item event payload for payment/settlement use.

## Payload Shape

Each item event should include:

- `order_id`
- `order_item_id`
- `product_id`
- `merchant_id`
- `buyer_user_id`
- `quantity`
- `unit_price_cents`
- `line_total_cents`

## Testing

- Keep tests in `services/orders/tests/`.
- Cover order creation, status updates, and outbox writes.
