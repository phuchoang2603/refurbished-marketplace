## Why

The marketplace runs its own Payment service against `tools/payment-gateway-simulator`, a page with pay/fail/expire buttons whose outcome callback reaches Payment unauthenticated. The payment-gateway-fraud-detection gateway (its `build-minimum-payment-gateway` change) now offers a hosted checkout page, an attempt-level fraud decision and signed outcome webhooks. The marketplace can use it as its payments platform and stop maintaining a second payment stack, while sending the checkout facts the gateway's fraud model is trained on.

## What Changes

- **BREAKING** Remove `services/payment`, `shared/proto/payment/v1`, the payment database, Kafka topics and Debezium `payment-outbox` connector, its chart values and release image, and `tools/payment-gateway-simulator` with its HTTPRoute and `pay` / `pay-dev` hosts.
- Products gain a fixed `category` and a `condition` grade (`excellent`, `good`, `fair`), chosen on the seller's create form and returned by batch reads.
- Checkout freezes a full gateway snapshot with the intent: line category and condition from the products batch read, seller and buyer account creation times from users, shipping address and return URL. Buyer email is not sent.
- After stock is reserved, Checkout sends one `POST /v1/payments` to the gateway with `Idempotency-Key = order_id` from a durable operation row, and stores the payment ID, hosted URL and expiry before reporting ready-to-pay. No merchant provisioning call.
- The gateway's notifier POSTs signed webhooks to web's in-cluster `POST /webhooks/payment-gateway`; web forwards body and signature to Checkout. Checkout verifies the HMAC and timestamp, deduplicates `event_id` in its inbox, and maps `payment.succeeded` to stock commit and `payment.expired` to stock release; `attempt.declined` does not end the checkout because the buyer can retry on the hosted page.
- At its payment deadline Checkout reads the gateway status once instead of cancelling a session: `SUCCEEDED` commits, `EXPIRED` releases, anything else is flagged for review while webhooks keep arriving.
- **BREAKING** Web drops `/callbacks/hosted-payment`; buyer return pages stay. Cart lines are removed after `SUCCEEDED` or `EXPIRED`.
- Prod only: Checkout calls the gateway Edge in-cluster with an API key and verifies webhooks with a signing secret, both from Doppler; the mesh provides mTLS transport in both directions. Dev has no gateway, so checkout submission there is refused with an unavailable message.

Non-goals: card entry inside the marketplace, merging gateway services into this repository, payment cancellation (the gateway has none), a dev gateway, refunds and multi-attempt payment UI in the marketplace, a marketplace traffic simulator.

## Capabilities

### New Capabilities

- `payment-gateway-integration`: Checkout's contract with the gateway — snapshot sources, one idempotent payment creation, verified webhooks and the deadline status read.

### Modified Capabilities

- `payment`: Removed; its responsibilities move to the gateway and `payment-gateway-integration`.
- `products`: Listings carry category and condition.
- `checkout`: Intent snapshot includes grading and account ages; payment step uses the gateway payment and its webhooks; deadline reads status instead of cancelling.
- `web`: Webhook forwarding route replaces the callback; seller form adds category and condition; checkout batch re-validation carries them; return and cart-removal rules use gateway statuses.
- `argocd-gitops`: Remove the simulator and hosted-payment URL requirements; Checkout gets gateway configuration in prod only.
- `istio-ambient-networking`: Gateway routes only the web hosts; the `pay` hosts leave the marketplace.
- `istio-resilience`: No retry of webhook POSTs; duplicate webhooks stay safe at Checkout.
- `external-secrets`: Payment Postgres credentials removed; gateway API key and webhook secret added.
- `distributed-tracing`: Payment outbox and callback tracing replaced by the webhook path.
- `structured-logging`: Webhook handling replaces payment-service logging rules.
- `github-actions-ci`: Payment removed from the shared fan-out map.

## Impact

- Code: delete `services/payment`, `shared/proto/payment/v1` and `tools/payment-gateway-simulator`; change `shared/proto/products/v1` (category, condition), `services/products`, `services/checkout` (snapshot, gateway client, operation rows, `HandleGatewayWebhook`, migration), `services/web` (seller form, webhook route), `shared/proto/checkout/v1`, `shared/messaging` payment event types.
- Infra: `infra/charts/kafka/values.yaml` payment topics and connector, `infra/charts/refurbished-marketplace` (payment and simulator workloads, HTTPRoutes, ExternalSecrets, checkout env, network policy admitting the gateway's notifier to web), Cloudflare tunnel hostnames for `pay` / `pay-dev`, `release-images.yml`, CI path filters.
- Dependencies: the gateway's `build-minimum-payment-gateway` change deployed in prod, with the marketplace integration configured there (API key, webhook URL `http://web.ecommerce.svc.cluster.local:8080/webhooks/payment-gateway`, signing secret, events `payment.succeeded`, `payment.expired`, `attempt.declined`).
- Docs: architecture, checkout flow, secrets and GitOps pages that describe the Payment service or simulator.
