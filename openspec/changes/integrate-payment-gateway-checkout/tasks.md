## 1. Contracts

- [x] 1.1 `shared/proto/products/v1`: add `category` and `condition` to `Product`, `CreateProductRequest` and `ProductCreated`; regenerate bindings
- [x] 1.2 `shared/proto/checkout/v1`: add category and condition to `CheckoutLine`, drop `buyer_email`, add `HandleGatewayWebhook(raw_body, signature_header)` with accepted/duplicate/rejected results; remove `CreatePaymentSession`, `CancelPaymentSession` and the payment result messages; regenerate bindings
- [x] 1.3 `shared/messaging`: remove payment event types and topics

## 2. Products

- [x] 2.1 Validate and persist category and condition against the fixed sets, return them on all reads and in `ProductCreated`

## 3. Checkout

- [x] 3.1 Migration `002`: rename `payment_session_id` → `gateway_payment_id`, `payment_return_url` → `hosted_checkout_url`, add `gateway_snapshot`, `gateway_expires_at`, `gateway_last_event_at`, create `checkout_gateway_operations`; update sqlc queries
- [x] 3.2 On `SubmitCheckout`, read seller and buyer `created_at` from users and freeze the gateway snapshot (lines with category and condition, account ages, shipping address, return URL) with the intent; ignore browser-supplied account times
- [x] 3.3 Gateway client (`POST /v1/payments` with `Idempotency-Key = order_id`, `GET /v1/payments/{id}`) with API key auth, timeouts and error classification; config `PAYMENT_GATEWAY_URL`, `PAYMENT_GATEWAY_API_KEY`, `PAYMENT_GATEWAY_WEBHOOK_SECRET`, `CHECKOUT_RETURN_BASE_URL`
- [x] 3.4 On `stock-reserved`, insert the create operation row in the same transaction instead of the `CreatePaymentSession` command
- [x] 3.5 Operation worker: claim due rows with `SKIP LOCKED`, resend the stored request with backoff, commit the gateway payment and READY_TO_PAY together, compensate on defect or preparation deadline
- [x] 3.6 `HandleGatewayWebhook`: verify signature and ±5 min tolerance, dedupe `gateway:<event_id>` in `checkout_inbox`, map `payment.succeeded` / `payment.expired` to commit / release in the same transaction as the outbox command, record `attempt.declined` time only
- [x] 3.7 Payment deadline: set from gateway expiry plus grace, replace `CancelPaymentSession` with one status read, write `payment_uncertain` when not terminal
- [x] 3.8 Refuse `SubmitCheckout` as `UNAVAILABLE` when no gateway URL is configured
- [x] 3.9 Logs and spans carry `order_id` and gateway payment ID, never the API key, secret, hosted URL token or raw webhook body

## 4. Web

- [x] 4.1 Seller create form: category and condition selects passed to products
- [x] 4.2 Checkout batch re-validation passes category and condition into `SubmitCheckout`; stop sending buyer email
- [x] 4.3 Add `POST /webhooks/payment-gateway` (64 KiB limit, raw body and `Gateway-Signature` forwarded to Checkout, 204/400/503 mapping), not attached to the public HTTPRoute; remove `/callbacks/hosted-payment` and web's Payment client
- [x] 4.4 Redirect to Checkout's hosted checkout URL; drop `PAYMENT_SVC_ADDR`, `HOSTED_PAYMENT_BASE_URL` and `HOSTED_PAYMENT_CALLBACK_BASE_URL`
- [x] 4.5 Order page and cart removal use SUCCEEDED / EXPIRED; render the checkout-unavailable message on `UNAVAILABLE`

## 5. Removal

- [x] 5.1 Delete `services/payment`, `shared/proto/payment/v1` and `tools/payment-gateway-simulator`
- [x] 5.2 Remove payment topics and the `payment-outbox` connector from `infra/charts/kafka/values.yaml`
- [x] 5.3 Remove payment and simulator workloads, `payment-db`, the `pay` HTTPRoute and payment gRPC retry entries from the marketplace chart and overlays
- [x] 5.4 Remove payment and simulator from `release-images.yml`, CI path filters and the shared fan-out map

## 6. Deployment

- [x] 6.1 Add the `checkout-payment-gateway` ExternalSecret (API key, webhook secret) and prod-only Checkout gateway values; leave dev without them
- [x] 6.2 Add a network policy admitting the gateway's notifier pods in `payment-gateway` to web's webhook port
- [x] 6.3 Remove payment Postgres keys from the Debezium connector secrets and document the new Doppler keys
- [x] 6.4 Update architecture, checkout flow, tracing, secrets and GitOps docs that describe the Payment service or simulator
- [ ] 6.5 Run the Rollout in prod: configure the integration, deploy, hand over the `pay` hostname, smoke test approve / decline-then-approve / expiry and confirm each webhook was accepted once
