## Context

Today the web service loads cart items, looks each product up through the products service, and builds a single mixed-merchant `CreateOrder` request that includes caller-supplied `merchant_id`, `unit_price_cents`, and `total_cents`. The orders service stores mixed-merchant line items, writes one `orders.item.created` outbox row per item, and the payment service creates one payment transaction per order item before emitting item-level outcome events back to orders.

That design works, but it makes the backend harder to follow than the marketplace domain requires. Merchant boundaries are reconstructed late, orders are not the unit of payment, and eventing is tied to line items instead of the actual merchant-scoped checkout work. The repo already carries merchant ownership on products, stores item-only cart state in Redis/Valkey, persists orders and payment state in PostgreSQL, and uses Kafka outbox/inbox patterns between orders and payment.

## Goals / Non-Goals

**Goals:**

- Make each persisted order belong to exactly one merchant.
- Make the core API shape merchant-scoped so later checkout orchestration can fan one cart out into multiple merchant-scoped orders.
- Move Kafka contracts from item-level to order-level so payment creates one transaction per order.
- Keep the implementation aligned with existing service ownership: products own product merchant identity, cart owns ephemeral cart state, orders own order persistence and event emission, payment owns payment execution.
- Keep the first implementation focused on service contracts, persistence, and eventing without changing the web edge.

**Non-Goals:**

- Introducing a dedicated merchant service, payout workflow, or merchant onboarding model.
- Creating a parent checkout aggregate or buyer-order hierarchy that groups the created merchant orders.
- Reworking product ownership semantics; products remain the canonical source of merchant ownership.
- Solving all cart staleness concerns beyond validating what is necessary at checkout time.
- Redesigning browser checkout orchestration or web-facing response models.

## Decisions

### Decision: Make orders single-merchant and leave checkout fan-out to a later edge redesign

The core API will move to a model where each persisted order belongs to exactly one merchant and mixed-merchant order creation is rejected. This stage does not define which edge or orchestrator fans a cart out into multiple calls.

Rationale:

- This keeps order invariants simple: one order, one merchant, one total, one payment lifecycle.
- It avoids introducing parent/child order state or partial aggregate statuses.
- It decouples the core model change from the web service, which the user intends to redesign later.

Alternatives considered:

- Keep one mixed-merchant order and introduce merchant suborders. Rejected because it adds aggregate order state, parent-child persistence, and more complicated failure handling.
- Add a new checkout service immediately. Rejected for the first iteration because it adds infrastructure and service boundaries before the simpler core model has stabilized.

### Decision: Require caller-supplied `merchant_id` in cart APIs and preserve it as cart state

Cart item write APIs will require callers to supply `merchant_id` together with `product_id` and `quantity`, and cart reads will return the same merchant-aware item shape. That keeps merchant routing data explicit in the core cart contract for whatever checkout orchestrator is introduced later.

Rationale:

- Cart becomes self-describing enough to group merchant buckets without recomputing ownership on every checkout pass.
- The upstream caller already determines which merchant context a cart line belongs to, so the cart service can stay a simple state store.
- Redis/Valkey cart state stays lightweight and focused on ephemeral checkout data.

Alternatives considered:

- Keep cart item state as `product_id` and `quantity` only, and always recompute merchant grouping from product lookups. Rejected because it preserves the late-bound merchant reconstruction that this change is trying to remove.
- Have the cart service look up products and derive `merchant_id` itself. Rejected because it would make cart depend on catalog authority instead of remaining a simple ephemeral state service.

### Decision: Accept caller-supplied merchant-scoped order creation and move merchant identity to the order level

Orders will store `merchant_id` on the order record itself. Order items will no longer need to carry a separate merchant identity as part of the core invariant. Order creation requests will become single-merchant requests that accept caller-supplied `merchant_id` and reject mixed-merchant items.

Rationale:

- The persistent model should mirror the domain invariant directly.
- Querying, indexing, and payment handoff all become simpler when merchant ownership is on the order.
- This reduces redundant merchant copies on every order item.
- The order service can trust the upstream cart-derived merchant grouping and avoid adding product lookups to the core order path in this stage.

Alternatives considered:

- Keep `merchant_id` only on order items even after fan-out. Rejected because it preserves a mixed-merchant data shape after the system stops allowing mixed-merchant orders.
- Recompute merchant ownership from products inside the orders service. Rejected for this stage because it adds a cross-service dependency and duplicates caller-owned grouping logic.

### Decision: Continue accepting caller-supplied unit prices in order creation

Order creation will continue to accept caller-supplied `unit_price_cents` from the upstream caller in this stage rather than verifying prices against the products service.

Rationale:

- This keeps the core order path simple and avoids introducing a new orders-to-products dependency during the merchant-scope refactor.
- It matches the same trust boundary being used for caller-supplied `merchant_id`.
- It keeps this change focused on order shape, persistence, and event granularity rather than broader pricing authority redesign.

Alternatives considered:

- Recompute or verify prices in the orders service against products. Rejected for this stage because it expands scope into cross-service validation and pricing authority concerns.

### Decision: Replace item-level Kafka contracts with order-level contracts

The orders service will emit one order-created event per created order. The payment service will consume that event, create one payment transaction per order, and emit one payment outcome event per order.

Rationale:

- Event granularity now matches the backend unit of work.
- Idempotency simplifies around `order_id` instead of `order_item_id`.
- Order status updates in the orders service become direct reflections of one payment outcome per order.

Alternatives considered:

- Preserve `orders.item.created` and aggregate in payment. Rejected because it keeps item-level complexity after the core model has changed.
- Emit both item-level and order-level events. Rejected for the first iteration because dual contracts would increase migration surface and codepath complexity.

## Risks / Trade-offs

- [Caller-supplied merchant or price data can diverge from catalog truth] -> Keep this stage focused on contract simplification; if cross-service verification becomes necessary later, add it as a separate change at the orchestrator or order boundary.
- [Edge orchestration is deferred while core APIs become stricter] -> Keep the change focused on rejecting mixed-merchant orders and on publishing the new order-level contracts; implement fan-out in a follow-up change.
- [Event contract migration breaks consumers] -> Migrate orders and payment together in one change and remove item-level topics only after both sides consume/emit the new order-level contracts.
- [Schema migration complexity around `merchant_id`] -> Add order-level merchant columns and new event payloads first, then remove redundant item-level fields after application code has switched over.
- [Cart now requires merchant metadata from callers] -> Make the contract explicit in proto, validation, and tests so downstream callers fail fast when they omit merchant context.

## Migration Plan

1. Extend cart contracts and cart state serialization to include `merchant_id` on cart items.
2. Add order-level `merchant_id` persistence and update order creation APIs to represent a single-merchant order request.
3. Introduce new order-level Kafka event names and payloads in shared proto and messaging definitions.
4. Update payment to consume order-level order-created events and emit order-level payment result events.
5. Update orders to consume the new payment result events and update one order per event.
6. Remove item-level event usage and any now-redundant order-item merchant plumbing after the new flow is active.

Rollback strategy:

- Roll back before the event-contract cutover by leaving item-level topics in place while new code is disabled.
- After cutover, rollback requires restoring the previous orders/payment binaries together because topic contracts change as a pair.

## Open Questions

None at this stage.
