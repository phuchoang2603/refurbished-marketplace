## Context

See `proposal.md` for motivation. The current checkout path in `services/web` performs three synchronous steps in sequence: create order, remove cart items from Redis, and create hosted payment session before redirecting the browser. The repository already uses transactional outbox and consumer inbox patterns for downstream order, inventory, and payment events, and `docs/order-placement.md` reflects that asynchronous choreography. The remaining failure mode is at the browser edge: a retry after partial synchronous failure can create a duplicate order or strand a pending order without a safe payment resume path.

Redis cart state is intentionally ephemeral. That means cart cleanup should not be elevated into a durable transactional participant alongside orders and payment persistence. The design should preserve the fast synchronous browser redirect path while making repeated checkout submits reconcile durable state instead of creating new state blindly.

## Goals / Non-Goals

**Goals:**

- Introduce a durable checkout-intent identity that lets the web edge and orders service treat retries as the same purchase attempt.
- Preserve one durable order and one durable hosted payment continuation per buyer checkout intent.
- Keep the buyer-facing happy path synchronous: submit checkout, ensure durable state, redirect to hosted payment.
- Keep outbox or inbox patterns for post-order lifecycle choreography and clarify where they do and do not apply.
- Make Redis cart cleanup non-critical to checkout success.

**Non-Goals:**

- Introduce XA, two-phase commit, or a new saga coordinator across Redis, orders PostgreSQL, and payment PostgreSQL.
- Rebuild the existing order, inventory, and payment event choreography into a different orchestration model.
- Turn cart state into a durable source of truth for commerce lifecycle.
- Add a new long-running reservation-expiry design beyond the existing payment-session expiry and downstream failure handling.

## Decisions

### Decision: Use checkout-intent idempotency at the browser edge and orders boundary

The browser edge will generate or preserve one checkout-intent key per submit intent, and `CreateOrder` will accept that key as part of its durable contract. Orders persistence will scope uniqueness to the buyer and the idempotency key so retries return the existing order instead of creating duplicates.

Rationale:

- The failure being addressed is buyer retry after a partially successful synchronous request.
- Transactional outbox does not answer whether a second HTTP submit is the same business attempt.
- A buyer-scoped idempotency contract prevents duplicate orders regardless of whether the retry happens before or after cart cleanup.

Alternatives considered:

- Derive idempotency from cart contents. Rejected because a buyer must still be able to legitimately rebuy the same cart later.
- Rely on order ID created at the web edge before persistence. Rejected because the current failure happens before the browser has a stable durable order identity to resume against.

### Decision: Keep hosted payment continuation anchored on order ID and refresh unusable sessions

Payment will continue using `order_id` as the durable payment anchor, but repeated session requests must distinguish between reusable unpaid sessions and expired or unusable sessions. Retries for the same unpaid order will either return the existing session metadata or refresh that continuation without requiring a new order.

Rationale:

- The payment domain already models one hosted session record per order and is close to the required shape.
- Retrying payment setup should resume the existing order lifecycle, not mint parallel orders or force the buyer to start over.
- Refreshing unusable session state is simpler than introducing a separate saga or browser-only resume cache.

Alternatives considered:

- Add a second payment-specific idempotency key unrelated to order ID. Rejected because the order is already the durable purchase anchor once created.
- Force the buyer to reopen payment only from a separate order page action. Rejected because checkout retry itself should be safe and reconciling.

### Decision: Treat Redis cart cleanup as a side effect, not a transactional requirement

The checkout flow will not require successful Redis cart removal before the buyer can continue to payment once durable order and payment continuation state exist. Cart cleanup can happen after durable state is ensured, or later from a paid-order flow, but it must not determine whether checkout succeeded.

Rationale:

- Redis cart state is explicitly ephemeral and should not gate durable commerce state.
- The buyer-facing failure in issue 35 is worse when cart removal succeeds before payment setup and the buyer loses an easy retry path.
- Moving cart cleanup out of the critical path aligns architecture with the repo's choice to keep cart separate from order truth.

Alternatives considered:

- Keep cart removal in the critical path and add compensation. Rejected because it still leaves checkout hostage to an ephemeral system and increases coordination complexity.
- Build a full choreography saga around order create, cart cleanup, and payment setup. Rejected because edge retries still require idempotency, and making Redis a first-class saga participant is not justified.

### Decision: Keep outbox and inbox focused on post-create lifecycle choreography

This change will preserve the existing event-driven split: synchronous edge reconciliation gets the buyer to one durable order and payment continuation, while outbox and inbox continue handling `orders.created`, inventory reservation, payment outcomes, and release or commit behavior asynchronously.

Rationale:

- The repo already has working service ownership and event contracts in this area.
- Outbox and inbox are still the right answer for durable cross-service lifecycle events.
- Separating edge reconciliation from domain choreography makes failure handling clearer and docs easier to reason about.

Alternatives considered:

- Replace the current downstream choreography with a centralized synchronous orchestrator. Rejected because it expands scope and weakens decoupling without solving buyer retries better.

## Risks / Trade-offs

- Conflicting reuse of the same checkout-intent key needs a clear error contract → Reject the request when the buyer reuses a key with different merchant, items, or total so accidental or malicious replay cannot silently change an existing order.
- Refreshing expired payment sessions changes previously simple get-or-return semantics → Keep refresh scoped to unpaid or unusable sessions only and document that paid or terminal orders do not reopen a new checkout path.
- Delaying cart cleanup may leave stale lines visible longer → Treat order and payment state as truth, and update browser UX so retries or return flows can steer the buyer to the pending order instead of relying on cart contents.

## Migration Plan

1. Extend the orders API, persistence, and validation to accept and enforce buyer-scoped checkout-intent idempotency.
2. Update the web checkout form and handler to preserve a checkout-intent key across retries and call orders and payment using reconciliation semantics.
3. Tighten payment hosted-session behavior to return reusable sessions and refresh expired or unusable sessions for unpaid orders.
4. Adjust cart cleanup behavior so Redis failures do not fail a buyer after durable state exists.
5. Update stable docs to show the separation between synchronous edge reconciliation and existing asynchronous lifecycle choreography.
6. Rollback path: stop sending the new web behavior only after orders and payment can tolerate the newer request contract; database uniqueness changes should be additive and safe to leave in place if the web edge temporarily falls back.
