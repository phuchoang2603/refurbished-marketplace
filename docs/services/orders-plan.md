# Orders Vertical Slice Plan

## Goal

Implement checkout order headers and per-item outbox events that drive inventory and payment asynchronously.

## Scope

- Service: `services/orders`
- Transport: gRPC
- Storage: PostgreSQL (`orders_db`)
- Events: `orders.item.created` (outbox) and consume `payment.item.succeeded` / `payment.item.failed` to move orders to paid/failed (`cmd/orders` + `KAFKA_BOOTSTRAP_SERVERS`; see `infra/charts/refurbished-marketplace/values.yaml`).

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

The outbox row `aggregate_id` should be the `product_id` for that item.

## Testing

- Keep tests in `services/orders/tests/`.
- Cover order creation, status updates, and outbox writes.
- Kafka end-to-end: `go test ./tests/...` exercises `KafkaPaymentResultHandler` + `shared/messaging` consumer. The test only subscribes to `payment.item.succeeded` because an empty dev broker may not have `payment.item.failed` yet; `cmd/orders` still consumes both topics in-cluster.
