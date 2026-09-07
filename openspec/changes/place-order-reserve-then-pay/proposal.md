## Why

Checkout today creates a pending order, optionally empties the cart, then redirects to hosted payment **before** stock is held. Kafka reserves asynchronously, so a buyer can pay (or sit on the simulator) for an unreserved SKU, and a retry after a partial BFF failure can mint a second order. Production practice is **reserve, then take payment**, with a stable checkout intent so retries resume instead of duplicating work.

## What Changes

- Treat **orders `PlaceOrder` / `CreateOrder`** as the checkout command: persist a merchant-scoped order keyed by buyer **`idempotency_key`**, then **`ReserveStock` over gRPC** to products **before** returning to web. Same buyer + key returns the existing order (and does not double-reserve).
- Web checkout: send the intent key → `PlaceOrder` → **get-or-create / refresh** hosted payment session → 303. **Do not** remove cart lines on this POST.
- Keep Kafka for **payment outcomes**: `payment.succeeded` commits reservations and marks the order paid; `payment.failed` / session expiry **releases** stock and fails the order. Cart clear is a **paid** side effect (web callback after successful webhook), not part of the command path.
- **BREAKING** (internal): products MUST expose a reserve RPC that orders is allowed to call (Cilium allow-list). `orders.created` consumers MUST treat reserve as already done when the gRPC path succeeded (idempotent no-op), not as a second reserve.
- No new checkout microservice.

## Capabilities

### New Capabilities

- None. Checkout stays a command inside `orders`, not a new bounded context.

### Modified Capabilities

- `orders`: idempotent place-order; reserve stock in the command path before success; failed reserve cancels/fails the order without a payment redirect.
- `products`: ReserveStock gRPC as the primary hold; keep commit/release on payment events; Kafka `orders.created` reserve stays idempotent only.
- `payment`: hosted session get-or-create and refresh-if-expired; outcomes still drive commit/release.
- `web`: checkout form intent key; PlaceOrder then session then redirect; no cart multi-remove on checkout POST; cart remove after paid webhook; resume-payment UX for pending reserved orders.
- `cart`: checkout MUST NOT require a successful multi-remove before payment; lines remain until paid (or documented resume if cookie/cart still holds them).
- `cilium-mesh-policy`: allow `orders` → `products` gRPC for ReserveStock.

## Impact

- Proto: `CreateOrder` / PlaceOrder `idempotency_key`; products `ReserveStock` (or equivalent) RPC.
- Services: `services/orders`, `services/products`, `services/payment`, `services/web` (checkout + webhook + order pages).
- Mesh: Cilium CNP documented pairs (`infra/charts` marketplace mesh policy).
- Docs: `docs/order-placement.md` sequence (durable docs, not this proposal).
- Tests when implementing (prompted later per repo rules): double PlaceOrder, last-unit reserve, session reattach, no cart drain on failed pay.
- Related: GitHub #35 (identity + reattach); this change **adds** reserve-before-redirect, which #35’s original text deferred.

## Non-goals

- New checkout service, KurrentDB, or replacing Kafka/outbox for payment outcomes.
- 2PC across cart Redis, orders Postgres, and payment.
- Prepaid wallet / sync `ProcessPayment` in the checkout HTTP request.
- HTTP waiting on `inventory.reserved` Kafka instead of ReserveStock gRPC.
- Gateway HTTP retries on checkout POST (still no dataplane retries until identity is proven in prod).
