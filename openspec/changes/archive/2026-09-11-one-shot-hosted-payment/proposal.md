## Why

Hosted payment still treats one order as a retryable payment handle: session refresh after expiry/failure, Resume payment on the order page, and a payment transaction that is born only after `inventory.reserved`. That is more state than checkout needs, and it withholds amount/merchant from the session until Kafka catches up—poor preparation for a later fraud gateway. Failures should be one-shot: a dead order, a new checkout, a new `order_id`.

## What Changes

- **BREAKING (browser):** Remove Resume payment (`POST /orders/{id}/pay`) and session refresh after FAILED/EXPIRED. An unpaid order that fails or expires stays terminal; the buyer places a **new** order to try again.
- Keep distinct hosted-session statuses: `PENDING`, `SUCCEEDED`, `FAILED`, `EXPIRED` (no collapse of expire into fail).
- **BREAKING (payment gRPC):** `CreateHostedPaymentSession` snapshots commerce facts at create time (`merchant_id`, `total_cents`, optional shipping, optional line items) and creates the **payment transaction in the same request**, not from `inventory.reserved`.
- Idempotent create for the same `order_id` while `PENDING` returns the existing session; a second create after FAILED/EXPIRED/SUCCEEDED does **not** mint a new session.
- `inventory.reserved` still acks/dedupes but MUST NOT create a second transaction; it may apply an already-terminal session outcome if the webhook/expiry raced.
- Expiry sweep still marks `EXPIRED` (distinct) and emits `payment.failed` so stock/order unwind.
- Pass through shipping on session create when the web edge supplies it (field already exists). **Do not** add checkout shipping UI in this change.
- Simulator stays a three-button stand-in; do not add fraud scoring here.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `payment`: Session is one-shot; transaction is created with the session; `inventory.reserved` no longer births the tx; refresh-after-expiry requirement is removed; EXPIRED stays distinct.
- `web`: Checkout still creates session + redirect; resume-pay requirement is removed; cancel return stays gone; callbacks still forward SUCCEEDED/FAILED/EXPIRED.
- `server-rendered-web`: Order detail no longer offers Resume payment.

## Impact

- `shared/proto/payment/v1`, `services/payment` (create, Kafka handler, expiry catch-up), `services/web` (checkout request fields, order actions/pages).
- Downstream Kafka `payment.succeeded` / `payment.failed` contracts unchanged; orders/products consumers should keep working.
- Empty billing remains until a later checkout-shipping/billing change.

## Non-goals

- Checkout shipping/billing form and address validation.
- Real payment-gateway product, signed session URLs, HMAC webhooks, device/IP fraud features.
- Changing reserve-then-pay choreography on web (CreateOrder → ReserveStock → hosted session).
