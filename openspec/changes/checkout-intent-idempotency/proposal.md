## Why

Checkout currently treats order creation, cart removal, and hosted payment session setup as one linear browser-edge flow without a durable checkout-attempt identity. The repo already uses outbox and inbox to protect asynchronous order, inventory, and payment events, but that does not prevent duplicate orders or stranded pending orders when a buyer retries after a synchronous edge failure.

## What Changes

- Add a durable checkout-intent identity to order creation so repeated submits or retries for the same buyer return the same merchant-scoped order instead of creating duplicates.
- Change the web checkout path from create-once semantics to reconcile-or-resume semantics: ensure the order exists, ensure a hosted payment session exists for that order, and redirect the buyer safely on retry.
- Tighten hosted payment session behavior so repeated requests for the same unpaid order return the stored session or refresh an expired one instead of leaving the buyer stranded.
- Move cart removal out of the critical success path for checkout by treating Redis cart state as ephemeral browser convenience rather than a transactional participant in durable commerce state.
- Extend stable architecture docs to describe the split between retry-safe browser-edge reconciliation and existing outbox or inbox-driven order lifecycle choreography.
- Do not introduce a full distributed transaction, XA, or a new saga coordinator around Redis cart, orders PostgreSQL, and payment PostgreSQL.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `web`: checkout behavior changes from a linear create-remove-create flow to retry-safe order and payment reconciliation that does not depend on cart mutation succeeding first.
- `orders`: order creation behavior changes to accept a buyer-scoped checkout intent key and return the existing order for repeated creates of the same intent.
- `payment`: hosted payment session behavior changes so retries for the same unpaid order reattach to a durable session and can refresh expired session state for browser resume flows.

## Impact

- Affected APIs include `shared/proto/orders/v1` order creation, `services/orders` persistence and validation, and `services/payment` hosted session semantics.
- Affected browser-edge code includes `services/web/internal/handlers/cart`, checkout form generation, and order or payment resume flows.
- Redis cart remains ephemeral and non-transactional; checkout success can no longer depend on synchronous cart cleanup completing first.
- Stable docs such as `docs/order-placement.md` will need updates to distinguish edge reconciliation from downstream event choreography.
