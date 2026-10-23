# payment-gateway-integration Specification

## Purpose

Define how Checkout uses the payment gateway as the marketplace's payments platform: one idempotent payment creation with the full checkout snapshot, signed outcome webhooks, and a deadline status read, without holding card data.

## Requirements

### Requirement: Idempotent gateway payment creation

After stock is reserved, Checkout SHALL create a gateway payment with one `POST /v1/payments` carrying `Idempotency-Key` equal to the order ID and the frozen checkout snapshot: order ID, seller and buyer IDs with their account creation times, total and currency, line items with product ID, category, condition, quantity and unit price, shipping address and an HTTPS return URL. No merchant provisioning call SHALL be made. The call SHALL run from a durable Checkout operation record and SHALL resend the identical request on retry. Checkout SHALL persist the returned payment ID, hosted checkout URL and expiry before exposing the payment destination. A gateway conflict for the order's key SHALL be treated as a defect, not retried with changed facts.

#### Scenario: Crash after gateway accepted the payment

- **WHEN** Checkout stops after the gateway created the payment but before Checkout recorded the response
- **THEN** the retried request returns the same payment and hosted URL, and only one gateway payment exists for the order

#### Scenario: Gateway unavailable

- **WHEN** payment creation keeps failing with unavailable or timeout errors until the preparation deadline
- **THEN** Checkout releases stock and finalizes the order as failed after Inventory confirms release, and since no hosted URL was exposed, any payment created by a lost response cannot be paid and expires at the gateway

### Requirement: Snapshot facts come from authoritative sources

The category and condition of each line SHALL come from the products batch read at checkout. Seller and buyer account creation times SHALL come from the users service when Checkout persists the intent, never from the browser. All snapshot facts SHALL be frozen with the checkout intent so retries send identical requests. Buyer email SHALL NOT be sent to the gateway.

#### Scenario: Browser supplies an account age

- **WHEN** a checkout submission includes a buyer account creation time
- **THEN** it is ignored and the users service value is used

### Requirement: Verified gateway webhooks

Checkout SHALL accept a gateway webhook only when its `Gateway-Signature` HMAC-SHA256 over `t "." body` matches the configured signing secret and its timestamp is within 5 minutes. It SHALL record each accepted `event_id` in its inbox together with the state change and outgoing Inventory command, so a repeated event ID is acknowledged without another effect. It SHALL match the event to a checkout by gateway payment ID and acknowledge events for unknown payments without effect. `payment.succeeded` SHALL be definitive success and `payment.expired` definitive failure; `attempt.declined` SHALL NOT change checkout state.

#### Scenario: Forged webhook

- **WHEN** a webhook arrives with a missing or wrong signature or a stale timestamp
- **THEN** it is rejected and the checkout does not change

#### Scenario: Declined attempt then success

- **WHEN** `attempt.declined` and then `payment.succeeded` arrive for the same payment
- **THEN** the checkout stays ready-to-pay after the decline and proceeds to stock commit after the success

#### Scenario: Redelivered success

- **WHEN** the same `payment.succeeded` event ID is delivered twice
- **THEN** Checkout commits stock once and acknowledges both deliveries

### Requirement: Payment deadline review

When a payment deadline (the gateway expiry plus a configured grace) passes without a definitive webhook, Checkout SHALL read `GET /v1/payments/{payment_id}` once and apply `SUCCEEDED` or `EXPIRED` as it would the matching webhook. Any other status or a failed read SHALL keep the checkout non-terminal, flag it for review and keep accepting webhooks.

#### Scenario: Every webhook retry failed

- **WHEN** the gateway expired the payment but its webhook never arrived
- **THEN** the deadline status read observes `EXPIRED` and Checkout releases stock

### Requirement: Gateway credentials and data handling

Checkout SHALL reach the gateway and receive webhooks over plaintext HTTP carried by the Istio ambient mesh, which provides STRICT mTLS, without application TLS or certificates. It SHALL still authenticate to the gateway with an integration API key, because mesh identity does not identify the integration, and SHALL verify webhook signatures, because they authenticate the payload. It SHALL hold the API key and webhook signing secret from cluster secrets, neither logged. The marketplace SHALL NOT collect, store or forward card numbers; card entry happens only on the gateway's hosted page. Logs SHALL NOT contain the API key, signing secret, checkout token or raw webhook bodies.

#### Scenario: Missing gateway configuration

- **WHEN** Checkout runs without a gateway endpoint configured
- **THEN** it refuses new checkout submissions as unavailable and creates no order or reservation
