## 1. Contract

- [x] 1.1 Extend `CreateHostedPaymentSessionRequest` with `merchant_id`, `total_cents`, and optional line items; keep `shipping_address`
- [x] 1.2 Regenerate payment protobufs

## 2. Payment persistence and service

- [x] 2.1 Add intent snapshot storage for line items (migration + sqlc); persist shipping on create as today
- [x] 2.2 Create hosted session and payment transaction in one DB transaction; require merchant and amount
- [x] 2.3 PENDING create is idempotent (same session id); terminal statuses reject without refresh
- [x] 2.4 Remove `RefreshHostedPaymentSession` query and call path
- [x] 2.5 Change `inventory.reserved` handler to require existing tx, inbox ack, and `ensureTerminalOutcomeForOrder` only
- [x] 2.6 Keep expiry sweep writing EXPIRED and emitting `payment.failed` (tx now always exists at expire)

## 3. Web

- [x] 3.1 Pass merchant, total cents, order lines, and shipping (empty until collected) from checkout `CreateHostedPaymentSession`
- [x] 3.2 Remove `POST /orders/{id}/pay`, `CanResumePayment`, and Resume payment UI
- [x] 3.3 Keep callbacks forwarding SUCCEEDED, FAILED, and EXPIRED

## 4. Tests that already exist

- [x] 4.1 Update payment Kafka/service tests for tx-at-create and no session refresh
- [x] 4.2 Update web checkout/order page tests for snapshot fields and no resume control
