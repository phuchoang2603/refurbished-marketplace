## Context

See proposal.md for motivation. Today `CreateHostedPaymentSession` writes `payment_intents` only; `payment_transactions` appear from `inventory.reserved`. `RefreshHostedPaymentSession` and `POST /orders/{id}/pay` retry the same order. Hosted statuses already include distinct `EXPIRED`. Shipping is stored on the intent but checkout does not collect it yet.

Reserve-then-pay on web (CreateOrder → ReserveStock → session) stays.

## Goals / Non-Goals

**Goals:**

- One payment attempt per `order_id`.
- Amount, merchant, optional shipping, optional line snapshot available as soon as the session exists.
- Keep `EXPIRED` vs `FAILED` in storage and callbacks; both still drive `payment.failed` for orders/products.

**Non-Goals:**

- Checkout address UI (pass-through only).
- Changing hosted URL query-string shape or simulator fraud logic.
- HMAC / gateway-owned session IDs.

## Decisions

### 1. Create intent + transaction in one DB transaction at session create

Web already knows `merchant_id` and `total_cents` after batch revalidate. Pass them (and optional shipping + line items) on `CreateHostedPaymentSession`. Persist intent snapshot columns / JSON and insert `payment_transactions` with `INITIALIZED` and `idempotency_key = order:{order_id}`.

**Alternative:** Keep tx birth on Kafka. Rejected: fraud/gateway snapshot would still wait on `inventory.reserved`, and webhook/expiry still need the “no tx yet” catch-up.

### 2. `inventory.reserved` becomes catch-up + inbox only

Handler: load intent/tx (must exist); if missing, fail the message (retry) rather than creating a tx from Kafka. Call existing `ensureTerminalOutcomeForOrder` so a webhook/expiry that landed first still emits outbox. Inbox ack unchanged.

**Alternative:** Drop the consumer. Rejected: still need a durable hook if webhook beats the create-path emit, and reservation remains the signal that stock is held.

### 3. One-shot create semantics

- No row → insert session + tx, 30-minute `expires_at`.
- PENDING (including unexpired or even past-expiry but not yet swept) → return existing metadata, do not rotate `payment_session_id`.
- SUCCEEDED / FAILED / EXPIRED → `FailedPrecondition` (or `AlreadyExists`); web checkout that hits this after a double-submit of a _new_ intent is a different `order_id`.

Same-form checkout retry still uses orders `checkout_intent` so it reuses the **order**, then Create session hits PENDING reuse.

### 4. Remove resume in web, keep expiry sweep

Delete `handleResumePayment`, `CanResumePayment`, Resume button. Expiry loop still sets `EXPIRED` and emits `payment.failed`. Simulator can still POST `EXPIRED`.

Gateway decline → `FAILED`. Timer or gateway abandon → `EXPIRED`. Downstream Kafka event is still `payment.failed` for both (orders already treat failure as fail-order).

### 5. Shipping

Keep proto `shipping_address`. Checkout may send empty until a later change. Do not block this change on address forms.

### 6. Line items

Add optional repeated `{product_id, quantity, unit_price_cents}` on create and store JSON on the intent for later fraud. Web should send the same lines it used for `CreateOrder`.

## Risks / Trade-offs

- [Buyer must re-checkout after expire/fail] → Accepted; each attempt is a new `order_id` for velocity features later.
- [Create session before Kafka reserve event] → Tx exists early; if reserve fails, web already `UpdateOrderStatus(FAILED)`. Payment row may sit INITIALIZED until expiry marks EXPIRED/`payment.failed`. Prefer: on reserve failure web already fails the order; optionally mark session FAILED in that path (web does not call payment today). **Mitigation:** expiry still cleans PENDING; products release on `payment.failed` or existing fail-order from web. If web failed the order without a payment event, products already released on reserve failure. Payment INITIALIZED leftover is ok until expiry.
- [Proto/web deploy skew] → Additive fields; payment rejects missing merchant/amount. Deploy payment then web.
- [Kafka create-tx tests] → Rewrite to assume tx exists.

## Migration Plan

1. Payment migration: JSON snapshot for lines if not using existing columns; merchant/amount already on `payment_transactions`.
2. Ship payment service (create path + Kafka).
3. Ship web (request fields, remove resume).
4. No data backfill required for empty shop; leftover PENDING sessions expire as today.

Rollback: revert web first (old web would not send merchant/amount → create fails). Not a soft rollback for resume UI.

## Open Questions

None that block implementation. Checkout shipping UI is a later change.
