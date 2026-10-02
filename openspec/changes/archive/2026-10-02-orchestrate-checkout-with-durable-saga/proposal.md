## Why

Web currently creates an order, reserves stock over gRPC, and creates a hosted payment session in one request, while Kafka independently delivers reservation and payment outcomes. The `inventory.reserved` consumer can run before the payment transaction commits; failures and stock release are not coordinated by one durable checkout state machine. Move checkout progression into a dedicated service so retries, timeouts, late callbacks, and compensation have explicit ownership.

## What Changes

- Add a dedicated Checkout service with its own PostgreSQL-backed saga, inbox/outbox, durable deadlines, and buyer-scoped checkout status. Redis/Valkey remains cart storage, not saga storage.
- **BREAKING**: Web submits one merchant-scoped, validated checkout snapshot and idempotency key, then displays pending progress until Checkout has committed a stock reservation and a hosted session is ready. Web no longer calls Orders, Inventory, and Payment sequentially to complete checkout.
- **BREAKING**: Orders accepts an idempotent checkout order command and changes order status only on Checkout's explicit finalization command; payment and reservation events no longer independently finalize it.
- **BREAKING**: Inventory accepts idempotent reserve, commit, and release commands, publishes correlated results, and never relies on web's synchronous reserve call or an `orders.created` reservation safety net for checkout.
- **BREAKING**: Payment creates a hosted session on Checkout's command after confirmed reservation, publishes readiness and terminal outcomes, and retires its `inventory.reserved` catch-up consumer. Web/gateway callbacks remain an ingress, not a second saga coordinator.
- Checkout waits for stock settlement before completing the order. It handles timeouts, duplicate/out-of-order events, uncertain gateway outcomes, and post-terminal payment success via reconciliation rather than silently dropping them.

## Capabilities

### New Capabilities

- `checkout`: Merchant-scoped saga submission, state transitions, durable timers, buyer progress, replay safety, compensation, and reconciliation.

### Modified Capabilities

- `web`: Accept and render asynchronous checkout progress; redirect only when a reserved order's hosted session is ready.
- `orders`: Create and finalize orders from idempotent checkout commands; stop independently applying inventory/payment outcome events.
- `inventory`: Reserve and settle stock from saga commands and report durable, correlated results.
- `payment`: Create a hosted session only after reservation and report gateway/expiry outcomes to Checkout without requiring an inventory-reserved catch-up.

## Impact

- New `services/checkout/` module, PostgreSQL database/migrations/sqlc, gRPC status/submit API, Kafka consumer, outbox connector, tracing and metrics, Helm deployment and mesh policy, and shared versioned protobuf/event contracts.
- Refactor `services/web/internal/handlers/cart/`, Orders/Inventory/Payment consumers and APIs, the hosted-payment simulator callback path, and `docs/order-placement.md`/architecture diagrams.
- Existing production data stays in place: legacy checkout/order/payment records and active sessions will not be migrated, drained, or retroactively reconciled. The saga owns only new checkouts; validate it with isolated synthetic data without resetting production. No new event store, Redis saga source of truth, real payment provider, or multi-merchant atomic checkout is in scope.
