## Why

The current checkout flow carries merchant ownership and pricing through the web edge into a single mixed-merchant order, then fans payment work out per order item. That makes the backend harder to reason about because merchant boundaries are reconstructed late, events are too fine-grained, and payment status is derived from item-level updates instead of one merchant-scoped order at a time.

This change simplifies the marketplace backend by making the core order and payment model merchant-scoped. Each resulting order becomes the unit for persistence, eventing, and payment lifecycle handling, while browser-edge orchestration is deferred to a later change.

## What Changes

- Change cart state to carry caller-supplied `merchant_id` alongside `product_id` and `quantity` so upstream flows can preserve merchant grouping explicitly.
- Change the core order contract so one order creation request represents exactly one caller-supplied merchant instead of a mixed-merchant checkout aggregate.
- Change the orders contract so each order belongs to exactly one merchant and order creation emits one order-level outbox event per created order instead of one event per order item.
- Change payment ingestion and outbox handling to operate on merchant-scoped order events and merchant-scoped payment transaction results rather than item-level events.
- Preserve product ownership of canonical `merchant_id` as a separate domain concern, but do not recompute merchant ownership inside cart or orders during this stage.
- Non-goals: changing `services/web` behavior, introducing a parent checkout aggregate, redesigning storefront UX, or adding merchant onboarding, payout, or authorization workflows.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cart`: cart state and cart-facing APIs now carry merchant-aware item data for checkout grouping.
- `orders`: orders become single-merchant records and emit order-level outbox events instead of item-level events.
- `payment`: payment consumes order-level events and emits order-level payment outcome events.

## Impact

- Affected code: `services/cart`, `services/orders`, `services/payment`, `shared/proto/cart/v1`, `shared/proto/orders/v1`, `shared/proto/payment/v1`, and shared Kafka event names/payloads.
- Affected persistence: cart state shape in Redis/Valkey, orders schema and queries, payment transaction shape and outbox/inbox handling.
- Affected async contracts: replace `orders.item.created`, `payment.item.succeeded`, and `payment.item.failed` with order-level equivalents.
- Affected internal APIs: cart item payloads, order creation contracts, and payment event payloads become merchant-scoped.
