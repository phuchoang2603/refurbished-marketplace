## ADDED Requirements

### Requirement: Shared OpenTelemetry bootstrap exports to VictoriaTraces

The repository SHALL provide a shared Go OpenTelemetry bootstrap that configures a tracer provider, W3C Trace Context propagation, and OTLP export to the platform VictoriaTraces backend used by Grafana.

#### Scenario: Service starts with tracing configured

- **WHEN** a marketplace service enables the shared tracing bootstrap with a VictoriaTraces OTLP endpoint
- **THEN** spans created by that service are exportable to VictoriaTraces for Grafana Explore

#### Scenario: W3C is the propagation format

- **WHEN** the shared tracing bootstrap configures propagators
- **THEN** it uses W3C `traceparent` / `tracestate` so app spans can join Istio mesh spans that use OpenTelemetry tracing

### Requirement: Sync path propagates one TraceId over HTTP and gRPC

Marketplace browser and gRPC hops on the checkout and hosted-payment callback paths SHALL continue a single W3C TraceId across process boundaries.

#### Scenario: Web continues or starts a trace and injects gRPC metadata

- **WHEN** `web` handles a browser or callback request and calls an internal gRPC service
- **THEN** it exports a server span and injects W3C trace context into the outgoing gRPC call metadata

#### Scenario: gRPC servers record server spans

- **WHEN** an instrumented gRPC server receives a request with W3C trace context
- **THEN** it creates a server span that continues the same TraceId

### Requirement: Outbox rows carry serialized span context

Orders, payment, and inventory outbox writes SHALL persist the active span context in a dedicated column in the same database transaction as the outbox event so Debezium can restore the trace after CDC.

#### Scenario: Outbox insert stores tracing context

- **WHEN** a service inserts an outbox row while a span is active
- **THEN** the row includes a `tracingspancontext` (or equivalently configured) field populated from the active context in the same transaction

#### Scenario: Schema exists for all three outboxes

- **WHEN** marketplace migrations for this change are applied
- **THEN** `orders_outbox`, `payment_outbox`, and `inventory_outbox` each have a tracing span context column

### Requirement: Debezium EventRouter emits Kafka traceparent

The Debezium outbox connectors SHALL restore span context from the outbox tracing field and emit W3C `traceparent` on Kafka records, with the OpenTelemetry SDK available on the Connect runtime image.

#### Scenario: Connect image includes OpenTelemetry

- **WHEN** the `connect-debezium` image is built for this change
- **THEN** it includes the OpenTelemetry API/SDK dependencies required for Debezium tracing integration

#### Scenario: EventRouter tracing maps the outbox field

- **WHEN** an outbox connector with EventRouter tracing enabled reads a new outbox row that has span context
- **THEN** the produced Kafka record includes a `traceparent` header continuing that TraceId

### Requirement: Consumers continue traces as child spans

Kafka consumers for marketplace domain events SHALL extract W3C context from message headers and create child spans of the upstream context (parent–child), not span-links-only, for the v1 visualization model.

#### Scenario: Inventory handles orders.created under parent context

- **WHEN** the inventory consumer processes `orders.created` with a `traceparent` header
- **THEN** it creates a child span under that TraceId visible in VictoriaTraces / Grafana

#### Scenario: Payment outbox consumer path continues context

- **WHEN** a consumer processes a payment outbox–routed event that carries `traceparent`
- **THEN** it creates a child span under the same TraceId

### Requirement: End-to-end checkout TraceId is verifiable

A staging checkout and hosted-payment callback SHALL produce a single connected TraceId spanning edge, domain services, outbox/Debezium, and consumers as documented.

#### Scenario: Checkout waterfall is connected

- **WHEN** an operator places an order through staging checkout after this change
- **THEN** Grafana Explore against VictoriaTraces shows one TraceId covering web → CreateOrder → outbox → Debezium → inventory handling

#### Scenario: Hosted payment callback is connected

- **WHEN** an operator completes a hosted-payment success or failure callback in staging
- **THEN** Grafana Explore shows one TraceId covering the callback → payment gRPC → payment outbox path as applicable

### Requirement: Tracing documentation

The repository SHALL document the end-to-end tracing architecture, TraceId joining rules (app + mesh), outbox/Debezium configuration, and Grafana verification steps.

#### Scenario: Contributor finds the tracing guide

- **WHEN** a contributor opens observability documentation after this change
- **THEN** they can follow steps to locate a checkout TraceId and interpret mesh vs app vs async spans
