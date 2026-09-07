## Context

See proposal.md for motivation. Today `services/web/internal/handlers/cart/checkout.go` calls `CreateOrder`, then cart multi-remove, then `CreateHostedPaymentSession`. Products reserve on Kafka `orders.created` (`HandleOrdersCreated`). Payment opens a session immediately and creates the payment **transaction** only after `inventory.reserved`. Commit/release already follow `payment.succeeded` / `payment.failed`.

Durable sequence docs live in `docs/order-placement.md` and MUST be updated when this ships (not in this change folder).

Orders has no products gRPC client. Cilium marketplace authorization is identity allow-lists (`openspec/specs/cilium-mesh-policy/spec.md`); **web** already calls products. Do not add `orders` → `products`.

## Goals / Non-Goals

**Goals:**

- Place-order persists the order; web holds stock over gRPC before hosted redirect.
- Web orchestrates: intent key, batch price re-validate, CreateOrder, ReserveStock, session get-or-create/refresh, 303.
- Kafka remains the outcome bus for commit/release and order paid/failed.
- Cart drain happens after paid, via the existing hosted-payment callback path in web (cart has no Kafka consumer).

**Non-Goals:**

- A checkout microservice or saga table beyond order + reservation + payment session rows.
- Dropping `orders.created` outbox in this change (keep it; products consume must be idempotent).
- Creating the payment transaction inside PlaceOrder (payment still keys the tx off `inventory.reserved`).
- Gateway retries on checkout POST.

## Decisions

### Decision: Web orchestrates CreateOrder then ReserveStock

Orders owns the order row and `orders.created` outbox. Products owns the hold. Web already talks to both, so checkout stays an edge command: persist, then reserve, then pay. Orders MUST NOT depend on products.

Retries: same intent returns the pending order; web calls ReserveStock again (idempotent). If reserve fails, web marks the order FAILED and rotates the checkout intent.

### Decision: Unique `(buyer_user_id, idempotency_key)` on orders

Web generates a UUID per checkout form submit (hidden field). Hash-of-cart is not the key (rebuy would collide).

On retry: return existing order if merchant and lines match; conflict otherwise. If the first attempt persisted the order but reserve failed, the order is FAILED and the same key is not payable; web rotates the intent.

### Decision: Products `ReserveStock` gRPC is the hold; Kafka is a safety net

`ReserveStock` performs the same transactional reserve + `inventory.reserved` outbox as `HandleOrdersCreated`. The Kafka consumer stays so existing connectors keep working and late `orders.created` deliveries no-op when reservations already exist (inbox + unique reservation per order/product).

Do not wait in HTTP for Debezium. Do not dual-reserve.

### Decision: Web still opens the hosted session after PlaceOrder returns

Payment stays the PSP adapter. Session PK remains `order_id`. Implement get-or-create and **refresh-if-expired** for unpaid orders so resume-payment and checkout retry share one path.

Payment transaction creation stays on `inventory.reserved`. After sync reserve, that event should follow quickly; keep existing expiry catch-up if the session expires before the tx row exists.

### Decision: Cart clear on successful webhook, not Kafka and not checkout POST

Cart is Redis behind web gRPC and has no Kafka consumer. Adding cart Kafka is out of scope. After `HandleGatewayWebhook` reports success, web multi-removes that order’s product IDs when `cart_id` is known. Failed/expired payment leaves lines in the cart so the buyer can retry with a **new** intent only after the previous order is failed and stock released—or resume the same pending order via resume-payment.

### Decision: Compensation stays event-driven

`payment.failed` / expiry still release stock and fail the order. Place-order does not call Release on payment failure (the buyer is gone). Place-order **does** require products not to leave a partial hold if ReserveStock fails mid-order (already true for Kafka reserve).

## Risks / Trade-offs

- [Checkout HTTP includes products ReserveStock] → Acceptable vs redirect-before-reserve; watch checkout timeout vs Gateway `backendRequest`.
- [Order row exists then reserve fails] → Web marks FAILED and rotates intent; do not redirect to hosted payment.
- [Webhook success but cart cookie missing] → Stock and order still settle; cart TTL expires leftovers. Document resume UX.
- [Double inventory.reserved] → Outbox/inbox uniqueness per order; payment create-tx already idempotent on `order_id`.

## Migration Plan

1. Proto + products ReserveStock + orders idempotency column; mesh allow-list.
2. Wire web checkout to ReserveStock after CreateOrder; keep Kafka consumer idempotent.
3. Strip any orders→products client; web already allowed to call products.
4. Payment session refresh-if-expired.
5. Update `docs/order-placement.md`.
6. Rollback: revert web to CreateOrder-without-reserve only if products ReserveStock is unused; do not mix redirect-before-reserve with a half-deployed unique key.

## Open Questions

None that change specs. Session refresh vs new `payment_session_id` after expiry is an implementation detail as long as the buyer gets a working hosted URL for the same `order_id`.
