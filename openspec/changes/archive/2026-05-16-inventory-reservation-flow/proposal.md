## Why

The current merchant-scoped checkout flow publishes only order-level `orders.created` events with no item lines, while the inventory spec still expects `orders.item.created`. That mismatch leaves inventory unable to reserve stock from the active Kafka contract and leaves payment able to proceed before stock is durably held.

## What Changes

- **BREAKING** Update the `orders.created` contract to include order item lines so downstream consumers can reserve stock without reintroducing item-level Kafka events.
- Change inventory from a product-counter-only consumer to an order-aware reservation consumer that records reservations idempotently per order and product.
- Insert inventory into the async checkout path so payment starts only after inventory confirms the full order reservation.
- Extend payment and orders to consume the new reservation outcomes and payment outcomes in the revised sequence.
- Keep scope limited to merchant-scoped order checkout, reservation commit/release, and the related Kafka contracts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `orders`: change the order-created outbox requirement to publish item lines and consume inventory reservation failure outcomes alongside payment outcomes.
- `payment`: change the payment trigger from direct `orders.created` consumption to successful inventory reservation for an order.
- `inventory`: change reservation requirements from item-level order events to order-level reservation ownership, reservation outcomes, and payment-driven commit/release.

## Impact

- Affected code in `services/orders`, `services/inventory`, `services/payment`, and `shared/proto` for event payloads and consumers.
- Inventory persistence will need durable reservation ownership data in addition to aggregate stock counters.
- Kafka topic usage and consumer sequencing will change across the checkout flow.
- Non-goals: cart changes, product pricing authority changes, and browser UX redesign.
