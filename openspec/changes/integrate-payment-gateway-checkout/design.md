## Context

Checkout runs the saga over Kafka: it writes commands to `checkout_outbox` (Debezium routes them), consumes results through `checkout_inbox`, and keeps durable deadlines on `checkouts`. Today the payment step sends `CreatePaymentSession` to the Payment service, waits for `PaymentSessionReady`, and on its payment deadline sends `CancelPaymentSession` as a fence. Payment owns sessions, callbacks and the `payment-outbox` connector; web forwards `POST /callbacks/hosted-payment` to it, and `tools/payment-gateway-simulator` renders the pay page on `pay` / `pay-dev`. Products store name, description, price and merchant only; users expose `created_at` through `GetUserByID`.

The gateway (payment-gateway-fraud-detection, `build-minimum-payment-gateway`) runs in the same prod cluster in namespace `payment-gateway`:

- Its JSON API is ClusterIP-only and authenticated by an integration API key from Edge configuration. `POST /v1/payments` takes the full snapshot and `Idempotency-Key`; merchant and customer IDs are derived, so there is no provisioning call.
- Its hosted page is public at `/checkout/{session_id}`.
- Its notifier POSTs `attempt.declined`, `payment.succeeded` and `payment.expired` JSON `{event_id, type, payment_id, order_reference, attempt_id, status, reason, occurred_at}` with `Idempotency-Key: <event_id>` and `Gateway-Signature: t=<unix>,v1=<hex HMAC-SHA256(secret, t "." body)>`, retrying about 12 times over 24 hours. Cluster-local HTTP URLs are allowed.
- It has no cancel command. A payment expires 15 minutes after creation, and never while an attempt outcome is unknown.
- Its fraud model is trained on line category (`phone`, `laptop`, `tablet`, `accessory`, `headphones`, `smartwatch`, `game_console`, `camera`), condition (`excellent`, `good`, `fair`) and account ages.

## Goals / Non-Goals

**Goals:**

- Checkout is the only marketplace component that talks to the gateway or holds its credentials.
- One idempotency rule on every hop: Checkout sends `order_id` as the creation key and deduplicates webhooks by `event_id` in its inbox.
- The payment step looks like every other saga step: a durable result committed with its state change.
- Live marketplace payments carry the same snapshot fields as the gateway's historical training data.

**Non-Goals:**

- Keeping the Payment service, its Kafka contract or the simulator as a fallback.
- Migrating in-flight Payment sessions (see Migration Plan).
- Showing per-attempt declines in marketplace pages; the hosted page shows them.

## Decisions

### Category and condition on products

`shared/proto/products/v1` `Product` and `CreateProductRequest` gain `category` and `condition` as strings validated against fixed sets in the products service, stored in the Mongo catalog document and in the `ProductCreated` event. The seller form adds two selects. A one-off migration sets `accessory` / `good` on existing documents. Strings rather than proto enums keep the values identical to the gateway contract without a mapping table.

Alternative: free-text condition in the description. Rejected because the gateway needs a categorical value.

### Snapshot frozen at intent

`SubmitCheckoutRequest` lines gain category and condition, set by web from the products batch read it already performs. Checkout calls users `GetUserByID` for the seller (the merchant's user) and buyer when it persists the intent, and stores both `created_at` values, the line facts, shipping address and return URL in `checkouts.gateway_snapshot` (JSON). Browser-supplied account times are not accepted. `buyer_email` is dropped from the request. Freezing everything at intent means every gateway retry sends identical bytes.

### Gateway payment creation from a checkout-owned operation table

The transition that records `stock-reserved` inserts a `checkout_gateway_operations` row (`checkout_id`, canonical request JSON built from `gateway_snapshot`, `attempts`, `next_attempt_at`, `status`) in the same transaction. A worker goroutine in Checkout claims due rows with `FOR UPDATE SKIP LOCKED`, sends `POST /v1/payments` with `Idempotency-Key: <order_id>`, and on success commits the payment ID, hosted URL and expiry together with the READY_TO_PAY transition. Transport errors and 5xx back off with jitter until the preparation deadline. Any 4xx, including 409, is a defect and moves the checkout to compensation, since the stored request never changes.

Alternatives: routing the call through `checkout_outbox` and a Kafka consumer adds a topic and connector for a request/response call; calling the gateway inside the `stock-reserved` handler holds the partition on a slow HTTP call and loses the response if the process dies before commit.

### Webhooks enter through web's in-cluster Service and are verified in Checkout

The marketplace integration's webhook URL is `http://web.ecommerce.svc.cluster.local:8080/webhooks/payment-gateway`. The route is not attached to the public Istio HTTPRoute; a network policy admits the gateway's `notifier` pods to web. Web reads at most 64 KiB and calls a new `checkout.v1.CheckoutService/HandleGatewayWebhook(raw_body, signature_header)`. Checkout:

1. Verifies the signature with a constant-time compare and a ±5 minute tolerance.
2. In one transaction inserts `gateway:<event_id>` into `checkout_inbox` (a duplicate means already processed), applies the transition and writes any Inventory command to `checkout_outbox`.

`payment.succeeded` follows the success path, `payment.expired` the release path, and `attempt.declined` only updates `gateway_last_event_at`. Result codes map to web responses: OK and duplicate → 204, bad signature or unknown body → 400, store unavailable → 503.

Alternatives: verifying in web would give web the signing secret and a second place that understands payment state; consuming the gateway's Kafka topic would couple the marketplace to the gateway's broker, schema and unauthenticated listener.

### Payment deadline reads status instead of fencing

The payment deadline becomes the gateway expiry plus a configurable grace (initially 2 minutes). Normally the terminal webhook has arrived by then. If not, Checkout calls `GET /v1/payments/{payment_id}` once: `SUCCEEDED` follows the success path and `EXPIRED` the release path; `PENDING` or an error writes a `payment_uncertain` exception row and leaves the checkout non-terminal, still accepting webhooks. A success after Checkout recorded expiry keeps the existing `contradictory_payment` exception path. `CancelPaymentSession` and the payment result topics go away.

### Checkout schema

Migration `002` renames `payment_session_id` to `gateway_payment_id` and `payment_return_url` to `hosted_checkout_url`, adds `gateway_snapshot`, `gateway_expires_at` and `gateway_last_event_at`, and creates `checkout_gateway_operations`. Buyer-visible status reads expose the hosted URL only in READY_TO_PAY.

### Configuration and environments

Set only in the prod overlay: `PAYMENT_GATEWAY_URL` (`http://edge.payment-gateway.svc.cluster.local:8080`), `PAYMENT_GATEWAY_API_KEY` and `PAYMENT_GATEWAY_WEBHOOK_SECRET` from the `checkout-payment-gateway` ExternalSecret, and `CHECKOUT_RETURN_BASE_URL`. With no URL configured, `SubmitCheckout` returns `UNAVAILABLE` and web shows that checkout is unavailable; no order or reservation is created. Checkout's egress to the gateway and the notifier's ingress to web leave the ambient mesh, so the API key and HMAC signature, not mesh identity, authenticate them.

## Risks / Trade-offs

- [Gateway down while stock is reserved] → The operation retries until the preparation deadline, then stock is released; buyers see a failed checkout rather than a stuck one.
- [Webhook delivery fails for hours] → The deadline status read resolves the checkout without webhooks.
- [Forged in-cluster webhook] → Signature check before any state read, body size limit, network policy limited to the notifier; forged requests cost one HMAC.
- [Existing listings lack grading] → Backfill defaults; sellers can edit later. Fraud features for those early orders are approximate.
- [Buyer pays after the marketplace gave up] → The gateway never expires a payment with an unknown attempt, and a late success raises the existing financial exception.
- [Dev loses a working checkout] → Accepted until a sandbox gateway integration exists.
- [Two repos must land in order] → Marketplace prod values switch only after the gateway is serving with the marketplace integration configured.

## Migration Plan

1. In the gateway, configure the marketplace integration (API key, webhook URL above, signing secret, subscribed events) and store the API key and signing secret in Doppler prod for both repos.
2. Deploy the products change and run the grading backfill.
3. Disable checkout submission in prod and wait until no checkout is in READY_TO_PAY or settling under the Payment service.
4. Deploy Checkout and web with migration `002`, the new ExternalSecret and prod gateway values.
5. Remove the Payment workload, database, Kafka topics, `payment-outbox` connector, simulator and `pay` HTTPRoute in the same release; hand the `pay` tunnel hostname to the gateway's hosted page.
6. Smoke test in prod: approve with `4242…`, decline then approve on the same page, and let one payment expire; confirm each webhook was accepted once.

Rollback is a revert of the release while no gateway payment is open; the dropped Payment database is not restored, so step 3's drain is what keeps rollback safe.
