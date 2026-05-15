## 1. Shared contracts

- [x] 1.1 Update `shared/proto/cart/v1/cart.proto` and generated code so cart items and add/set item requests require caller-supplied `merchant_id`.
- [x] 1.2 Update `shared/proto/orders/v1/orders.proto` and related generated code so order creation is single-merchant, requires caller-supplied `merchant_id`, and order responses carry order-level `merchant_id`.
- [x] 1.3 Replace the order item created and item payment payload definitions with order-level event payloads under `shared/proto/orders/v1` and `shared/proto/payment/v1`.
- [x] 1.4 Update `shared/messaging/event_types.go` and any Kafka bootstrap wiring to use the new order-level topic names.

## 2. Cart service

- [x] 2.1 Update cart service state models, validation, and gRPC handlers to require, persist, and return merchant-aware cart items.
- [x] 2.2 Update cart service tests to cover rejecting missing `merchant_id` values and handling merchant-aware cart item reads and writes.

## 3. Orders service

- [x] 3.1 Add order-level `merchant_id` persistence in `services/orders/db/migrations` and regenerate sqlc queries/models.
- [x] 3.2 Update orders service types, validation, and gRPC handlers so one create-order request represents exactly one merchant-scoped order with caller-supplied `merchant_id`.
- [x] 3.3 Replace per-item outbox creation with one order-level outbox row per created order and update Kafka payload generation.
- [x] 3.4 Update orders Kafka consumers to process order-level payment success and failure events.
- [x] 3.5 Update orders service and Kafka tests for single-merchant validation, order-level outbox emission, and order-level payment status updates.

## 4. Payment service

- [x] 4.1 Update payment persistence and service logic to create one payment transaction per order keyed by `order_id`.
- [x] 4.2 Update payment Kafka consumers to ingest the new order-created event and write inbox records against the order-level contract.
- [x] 4.3 Update payment outbox emission to publish order-level payment success and failure events.
- [x] 4.4 Update payment service and Kafka tests for order-level transaction creation, deduplication, and outcome events.

## 5. Cutover and verification

- [x] 5.1 Remove obsolete item-level event handling and redundant order-item merchant plumbing once all consumers use the new order-level flow.
- [x] 5.2 Run targeted service tests for cart, orders, and payment, then fix any contract mismatches revealed by the new merchant-scoped core flow.
