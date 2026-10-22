## MODIFIED Requirements

### Requirement: Sensitive material is not logged in full

Structured logging helpers and access-log middleware SHALL NOT emit full payment card data, passwords, gateway API keys, webhook signing secrets, checkout tokens, or raw payment gateway payloads in log attributes.

#### Scenario: Payment callback path spot-check

- **WHEN** gateway webhook handling or gateway client paths log errors
- **THEN** logs do not contain full card numbers, CVV, secrets, checkout tokens, or complete raw gateway request/response bodies

### Requirement: Checkout hot paths emit domain fields

Orders, products (inventory), and checkout services SHALL emit structured slog lines on checkout hot paths that include domain identifiers such as `order_id` (and related `merchant_id` / `buyer_user_id` / outcome fields when applicable), using context so `trace_id` is present when a span is active.

#### Scenario: Order create logs order_id

- **WHEN** CreateOrder commits successfully
- **THEN** an Info log includes `order_id`, `buyer_user_id`, and `merchant_id`

#### Scenario: Inventory reservation outcome is logged

- **WHEN** products handles `orders.created` and reserves stock or emits reservation_failed
- **THEN** a structured log includes `order_id` and indicates reserved vs failed

#### Scenario: Payment terminal outcome is logged

- **WHEN** checkout applies a verified gateway webhook or status read
- **THEN** a structured log includes `order_id`, the gateway payment ID, and the payment status / event type without card or raw gateway payloads
