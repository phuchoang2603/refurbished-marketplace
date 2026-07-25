## ADDED Requirements

### Requirement: Shared JSON slog bootstrap with OTEL correlation

The repository SHALL provide a `shared/observe/log` Go module that configures slog with a JSON handler writing to stdout, attaches a default `service` attribute, and injects `trace_id` and `span_id` from the active OpenTelemetry span when logging with a context that has a valid span.

#### Scenario: Service initializes structured logging

- **WHEN** a marketplace service starts and initializes `shared/observe/log` with its service name
- **THEN** subsequent Info/Error logs are emitted as JSON lines to stdout including `service`

#### Scenario: Log line carries TraceId from active span

- **WHEN** code logs with a context that has a valid OTEL span
- **THEN** the JSON log line includes `trace_id` and `span_id` matching that span

#### Scenario: Log line omits TraceId without span

- **WHEN** code logs without a valid OTEL span on the context
- **THEN** the JSON log line is still valid and does not invent a fake `trace_id`

### Requirement: Production paths use structured logging

Marketplace service entrypoints and `shared/runtime` production logging SHALL use `shared/observe/log` instead of stdlib `log.Printf` / text loggers for operational messages.

#### Scenario: Runtime bootstrap logs are JSON

- **WHEN** `shared/runtime` emits startup, tracing status, or server listen messages
- **THEN** those messages are slog JSON lines with the service attribute set

#### Scenario: Service main and kafka entrypoints migrated

- **WHEN** a marketplace service `main` or kafka consumer entrypoint logs operational events or fatals
- **THEN** it uses slog (JSON) rather than stdlib `log.Printf` / `log.Fatal` text output

### Requirement: gRPC access logs are structured

Instrumented gRPC servers SHALL emit structured access logs for unary (or equivalent) RPCs including method, status code, duration, and `trace_id` when a span is active.

#### Scenario: Successful unary RPC is logged

- **WHEN** a gRPC unary handler completes successfully
- **THEN** a JSON access log includes the gRPC method, OK status code, duration, and `trace_id` when tracing is active

#### Scenario: Failed unary RPC is logged

- **WHEN** a gRPC unary handler returns a non-OK status
- **THEN** a JSON access log includes the method, status code, duration, and `trace_id` when tracing is active

### Requirement: Kafka consumer logs are structured

Kafka consumer handle and error paths in `shared/messaging` SHALL emit structured logs that include `topic`, `partition`, and `offset`, plus `trace_id` when a consumer span is active.

#### Scenario: Handler error includes Kafka coordinates

- **WHEN** a Kafka message handler returns an error
- **THEN** the error log is JSON and includes `topic`, `partition`, `offset`, and `trace_id` when a span is active

### Requirement: Sensitive material is not logged in full

Structured logging helpers and access-log middleware SHALL NOT emit full payment card data, passwords, or raw payment gateway payloads in log attributes.

#### Scenario: Payment callback path spot-check

- **WHEN** hosted-payment callback or payment processing paths log errors
- **THEN** logs do not contain full card numbers, CVV, or complete raw gateway request/response bodies

### Requirement: Logging documentation

The repository SHALL document structured logging field conventions, VictoriaLogs LogSQL examples filtered by `service`, `trace_id`, and checkout domain fields such as `order_id`, and Trace → logs usage in `docs/observability.md`.

#### Scenario: Contributor finds logging guide

- **WHEN** a contributor opens observability documentation after this change
- **THEN** they can find the field table (including domain hot-path fields), a LogSQL example joining by `trace_id` or `order_id`, and Trace → logs steps

### Requirement: Checkout hot paths emit domain fields

Orders, products (inventory), and payment services SHALL emit structured slog lines on checkout hot paths that include domain identifiers such as `order_id` (and related `merchant_id` / `buyer_user_id` / outcome fields when applicable), using context so `trace_id` is present when a span is active.

#### Scenario: Order create logs order_id

- **WHEN** CreateOrder commits successfully
- **THEN** an Info log includes `order_id`, `buyer_user_id`, and `merchant_id`

#### Scenario: Inventory reservation outcome is logged

- **WHEN** products handles `orders.created` and reserves stock or emits reservation_failed
- **THEN** a structured log includes `order_id` and indicates reserved vs failed

#### Scenario: Payment terminal outcome is logged

- **WHEN** payment applies a gateway webhook terminal success or failure
- **THEN** a structured log includes `order_id` and the payment status / event type without card or raw gateway payloads
