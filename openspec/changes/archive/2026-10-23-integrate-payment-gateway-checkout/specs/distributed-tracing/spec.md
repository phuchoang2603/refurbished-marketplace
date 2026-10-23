## MODIFIED Requirements

### Requirement: Sync path propagates one TraceId over HTTP and gRPC

Marketplace browser and gRPC hops on the checkout and gateway webhook paths SHALL continue a single W3C TraceId across process boundaries.

#### Scenario: Web continues or starts a trace and injects gRPC metadata

- **WHEN** `web` handles a browser or webhook request and calls an internal gRPC service
- **THEN** it exports a server span and injects W3C trace context into the outgoing gRPC call metadata

#### Scenario: gRPC servers record server spans

- **WHEN** an instrumented gRPC server receives a request with W3C trace context
- **THEN** it creates a server span that continues the same TraceId

### Requirement: Outbox rows carry serialized span context

Orders, inventory, and checkout outbox writes SHALL persist the active span context in a dedicated column in the same database transaction as the outbox event so Debezium can restore the trace after CDC.

#### Scenario: Outbox insert stores tracing context

- **WHEN** a service inserts an outbox row while a span is active
- **THEN** the row includes a `tracingspancontext` (or equivalently configured) field populated from the active context in the same transaction

#### Scenario: Schema exists for all three outboxes

- **WHEN** marketplace migrations for this change are applied
- **THEN** `orders_outbox`, `inventory_outbox`, and `checkout_outbox` each have a tracing span context column

### Requirement: Consumers continue traces as child spans

Kafka consumers for marketplace domain events SHALL extract W3C context from message headers and create child spans of the upstream context (parent–child), not span-links-only, for the v1 visualization model.

#### Scenario: Inventory handles orders.created under parent context

- **WHEN** the inventory consumer processes `orders.created` with a `traceparent` header
- **THEN** it creates a child span under that TraceId visible in HyperDX

#### Scenario: Payment outbox consumer path continues context

- **WHEN** Inventory consumes a stock commit or release command that Checkout wrote to its outbox while handling a gateway webhook
- **THEN** it creates a child span under the same TraceId

### Requirement: End-to-end checkout TraceId is verifiable

A checkout and gateway webhook SHALL produce a single connected TraceId spanning web, domain services, outbox/Debezium, and consumers as documented. Kafka Connect SHALL export its spans to the same platform agent. Mesh proxy spans are not required for verification.

#### Scenario: Checkout waterfall is connected

- **WHEN** an operator places an order through checkout in an environment with the gateway enabled
- **THEN** HyperDX shows one TraceId covering web → CreateOrder → outbox → Debezium → inventory handling, including DB child spans where those services query Postgres

#### Scenario: Hosted payment callback is connected

- **WHEN** the gateway delivers a `payment.succeeded` webhook
- **THEN** HyperDX shows one TraceId covering web's webhook route → Checkout webhook handling → checkout outbox → Debezium → Inventory commit

#### Scenario: Mesh proxy services are absent from the waterfall

- **WHEN** an operator opens a checkout TraceId after mesh tracing resources are removed
- **THEN** the waterfall does not include `ecommerce-ingress` or `ecommerce-waypoint` spans
