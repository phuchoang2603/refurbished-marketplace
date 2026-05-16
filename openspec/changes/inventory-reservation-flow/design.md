## Context

The active checkout contract is merchant-scoped and order-level: orders emits one `orders.created` event per order and payment currently consumes that event directly. The current `OrderCreated` payload includes only `order_id`, `buyer_user_id`, `merchant_id`, and `total_cents`, while the inventory spec still describes consumption of `orders.item.created` keyed by `product_id`.

The inventory service implementation also reflects the older shape. Its gRPC and SQL paths mutate `available_qty` and `reserved_qty` by `product_id`, but they do not persist reservation ownership by `order_id` or track message-level idempotency for order reservations. As a result, inventory cannot safely reserve, commit, or release stock from the current Kafka flow without additional reservation state.

This change spans `orders`, `inventory`, `payment`, `shared/proto`, and service-local persistence. It changes event payloads and consumer sequencing but stays within the existing Kafka-based eventual consistency model.

## Goals / Non-Goals

**Goals:**

- Keep Kafka order-level rather than reintroducing item-level topics.
- Enrich `orders.created` with the item lines inventory needs to reserve stock for an order.
- Make inventory the owner of durable reservation state per `order_id` and `product_id`.
- Gate payment on successful inventory reservation for the full order.
- Preserve payment-driven commit and release of reserved stock.

**Non-Goals:**

- Redesign cart or browser checkout flows.
- Rework product pricing authority or add catalog revalidation in orders.
- Introduce synchronous cross-service orchestration for checkout.
- Add mixed-merchant checkout fan-out beyond the current merchant-scoped order model.

## Decisions

### Decision: Enrich `orders.created` instead of reviving item-level order events

`orders.created` will remain the single order-created Kafka contract, but its payload will expand to include the order item lines needed for downstream reservation.

Rationale:

- It preserves the repo's existing shift toward one event per order.
- It gives inventory enough context to reserve the entire order without reconstructing item state from other services.
- It avoids dual order-created contracts that would complicate migration and testing.

Alternatives considered:

- Restore `orders.item.created`. Rejected because it undoes the order-level contract simplification already adopted.
- Emit both order-level and item-level events. Rejected because it increases migration surface and duplicates checkout semantics.
- Add a separate `inventory.reservation.requested` event emitted by orders. Rejected for this stage because it duplicates the order-created payload while not reducing coupling materially.

### Decision: Inventory will persist reservation ownership per order item line

Inventory will add durable reservation records keyed by `order_id` and `product_id` in addition to the existing aggregate stock counters. The reservation consumer will record the Kafka inbox message, reserve all order lines idempotently, and emit a reservation outcome event.

Rationale:

- Aggregate counters alone cannot safely answer which order owns reserved units.
- Payment outcomes need a stable reservation identity to commit or release the correct stock on retries.
- A reservation ledger gives inventory a place to model reservation status without moving ownership into another service.

Alternatives considered:

- Keep only `available_qty` and `reserved_qty`. Rejected because duplicate delivery and asynchronous compensation would be ambiguous per order.
- Store reservation ownership in orders. Rejected because reservation state belongs to inventory's persistence boundary.

### Decision: Introduce minimal inventory reservation outcome events and gate payment on success

Inventory will emit order-level reservation outcomes, with `inventory.reserved` for successful full-order reservation and `inventory.reservation_failed` when the order cannot be fully reserved. The failure event will stay minimal in the first iteration and will not introduce structured failure taxonomies beyond the information needed to identify the order. Payment will consume `inventory.reserved` instead of `orders.created`. Orders will consume `inventory.reservation_failed` and existing payment outcomes.

Rationale:

- Payment should not charge until stock is durably held.
- A separate reservation outcome gives orders a direct signal when an order cannot proceed because of inventory.
- Keeping outcomes order-level aligns with the existing unit of work and current order status model.

Alternatives considered:

- Let payment continue consuming `orders.created` and race inventory. Rejected because payment could succeed for out-of-stock orders.
- Have payment poll inventory synchronously. Rejected because it departs from the current Kafka integration pattern.

### Decision: Reuse order-level failure handling in orders for inventory reservation failures

Orders will treat `inventory.reservation_failed` as an order-level failure signal and update the order to a failed state, rather than introducing a separate order status in this change.

Rationale:

- It keeps the status model small while still allowing the order to stop progressing.
- It limits scope to the event sequencing and reservation changes that unblock implementation.

Alternatives considered:

- Add a dedicated out-of-stock order status. Rejected for this stage because it expands UI and API surface beyond what is necessary to start implementation.

### Decision: Exclude reservation expiry from the first iteration

This change will not add reservation expiry timers or background expiration handling. The initial implementation will rely on inventory reservation success followed by payment-driven commit or release.

Rationale:

- It keeps the first implementation focused on the current checkout path and Kafka contract migration.
- The existing gap is missing reservation ownership and event sequencing, not timeout orchestration.
- Expiry introduces additional scheduling and state-transition concerns that are not required to start implementation.

Alternatives considered:

- Add reservation TTL and automatic expiry now. Rejected because it expands the change beyond the minimum flow needed to safely reserve before payment.

## Risks / Trade-offs

- [Breaking `orders.created` payload compatibility] -> Update `orders`, `inventory`, and `payment` in one coordinated change and regenerate shared protobufs together.
- [Partial reservation across multi-line orders] -> Reserve within inventory-owned transactional logic and emit success only after all lines are durably reserved.
- [Duplicate Kafka delivery creates duplicate reservations] -> Record inbox messages and enforce reservation uniqueness by `order_id` and `product_id`.
- [Inventory failure semantics become too coarse with a generic failed order state] -> Keep failure reasons in event payloads and persistence so a later change can split statuses if needed.
- [Reservation records drift from aggregate counters] -> Keep ledger writes and stock counter mutations in the same inventory transaction.

## Migration Plan

1. Extend the shared order-created protobuf to include item lines and add inventory reservation outcome protobufs.
2. Add inventory reservation persistence, inbox/outbox support if missing, and consumer/producer paths for reservation outcomes.
3. Update payment to consume `inventory.reserved` and keep emitting `payment.succeeded` and `payment.failed`.
4. Update orders to publish the enriched `orders.created` payload and consume `inventory.reservation_failed` in addition to payment outcomes.
5. Update stable docs in `docs/` to reflect the new flow after the code path is in place.

Rollback strategy:

- Before deployment cutover, keep the new consumers disabled or undeployed while shared protobuf changes are staged.
- After cutover, rollback requires restoring the coordinated `orders`, `inventory`, and `payment` binaries together because the checkout event sequence changes across all three services.

## Open Questions

None at this stage.
