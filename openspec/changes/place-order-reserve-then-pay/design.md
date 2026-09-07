## Context

See proposal.md for motivation. Today `services/web/internal/handlers/cart/checkout.go` calls `CreateOrder`, then cart multi-remove, then `CreateHostedPaymentSession`. Products reserve on Kafka `orders.created` (`HandleOrdersCreated`). Payment opens a session immediately and creates the payment **transaction** only after `inventory.reserved`. Commit/release already follow `payment.succeeded` / `payment.failed`.

Durable sequence docs live in `docs/order-placement.md` and MUST be updated when this ships (not in this change folder).

Orders has no products gRPC client today. Cilium marketplace authorization is identity allow-lists (`openspec/specs/cilium-mesh-policy/spec.md`); a new `orders` → `products` hop MUST be added there and in the chart CNPs.

## Goals / Non-Goals

**Goals:**

- Place-order is the command: persist + gRPC reserve, then return.
- Web is a thin edge: intent key, batch price re-validate, PlaceOrder, session get-or-create/refresh, 303.
- Kafka remains the outcome bus for commit/release and order paid/failed.
- Cart drain happens after paid, via the existing hosted-payment callback path in web (cart has no Kafka consumer).

**Non-Goals:**

- A checkout microservice or saga table beyond order + reservation + payment session rows.
- Dropping `orders.created` outbox in this change (keep it; products consume must be idempotent).
- Creating the payment transaction inside PlaceOrder (payment still keys the tx off `inventory.reserved`).
- Gateway retries on checkout POST.

## Decisions

### Decision: Host PlaceOrder in orders, not a new service

Orders already owns the order row and `orders.created` outbox. Adding a products gRPC client there is cheaper than a fourth runtime.

Alternatives considered: web calling ReserveStock itself (splits command across two RPCs from the edge; retries can create an order without reserve unless web is perfect). New checkout service (rejected: relocates the same three steps).

### Decision: Unique `(buyer_user_id, idempotency_key)` on orders

Web generates a UUID per checkout form submit (hidden field). Hash-of-cart is not the key (rebuy would collide).

On retry: return existing order if merchant and lines match; conflict otherwise. If the first attempt persisted the order but reserve failed, retry with the same key MUST resume reserve (or fail the order cleanly) rather than insert a new row.

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

- [PlaceOrder latency includes products gRPC + row locks] → Acceptable vs redirect-before-reserve; watch checkout timeout vs Gateway `backendRequest`.
- [Order row exists then reserve fails] → Same intent retry must resume; do not leave the buyer with a payable unreserved order.
- [Webhook success but cart cookie missing] → Stock and order still settle; cart TTL expires leftovers. Document resume UX.
- [Mesh deny orders→products] → Ship CNP with the app change; fail closed in enforce mode.
- [Double inventory.reserved] → Outbox/inbox uniqueness per order; payment create-tx already idempotent on `order_id`.

## Migration Plan

1. Proto + products ReserveStock + orders idempotency column; mesh allow-list.
2. Wire PlaceOrder to ReserveStock; keep Kafka consumer idempotent.
3. Switch web checkout; stop checkout multi-remove; add callback multi-remove + resume-payment.
4. Payment session refresh-if-expired.
5. Update `docs/order-placement.md`.
6. Rollback: revert web to CreateOrder-without-reserve only if products ReserveStock is unused; do not mix redirect-before-reserve with a half-deployed unique key.

## Open Questions

None that change specs. Session refresh vs new `payment_session_id` after expiry is an implementation detail as long as the buyer gets a working hosted URL for the same `order_id`.
