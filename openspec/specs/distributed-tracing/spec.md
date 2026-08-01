# distributed-tracing Specification

## Purpose

TBD - created by archiving change add-e2e-distributed-tracing. Update Purpose after archive.

## Requirements

### Requirement: Shared OpenTelemetry bootstrap exports to VictoriaTraces

The repository SHALL provide a shared Go OpenTelemetry bootstrap under `shared/observe/trace` that configures a tracer provider, W3C Trace Context propagation, and OTLP export to the platform VictoriaTraces backend used by Grafana.

#### Scenario: Service starts with tracing configured

- **WHEN** a marketplace service enables the `shared/observe/trace` bootstrap with a VictoriaTraces OTLP endpoint
- **THEN** spans created by that service are exportable to VictoriaTraces for Grafana Explore

#### Scenario: W3C is the propagation format

- **WHEN** the shared tracing bootstrap configures propagators
- **THEN** it uses W3C `traceparent` / `tracestate` so TraceIds continue across HTTP, gRPC, and Kafka hops between marketplace services

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

A staging checkout and hosted-payment callback SHALL produce a single connected TraceId spanning web, domain services, outbox/Debezium, and consumers as documented. Mesh proxy spans are not required for verification.

#### Scenario: Checkout waterfall is connected

- **WHEN** an operator places an order through staging checkout after this change
- **THEN** Grafana Explore against VictoriaTraces shows one TraceId covering web → CreateOrder → outbox → Debezium → inventory handling, including DB child spans where those services query Postgres

#### Scenario: Hosted payment callback is connected

- **WHEN** an operator completes a hosted-payment success or failure callback in staging
- **THEN** Grafana Explore shows one TraceId covering the callback → payment gRPC → payment outbox path as applicable

#### Scenario: Mesh proxy services are absent from the waterfall

- **WHEN** an operator opens a checkout TraceId after mesh tracing resources are removed
- **THEN** the waterfall does not include `ecommerce-ingress` or `ecommerce-waypoint` spans

### Requirement: Tracing documentation

The repository SHALL document the end-to-end tracing architecture, TraceId joining rules for application and async hops, outbox/Debezium configuration, operation-centric naming expectations, DB/Redis child spans, and Grafana verification steps that do not depend on Istio proxy spans.

#### Scenario: Contributor finds the tracing guide

- **WHEN** a contributor opens observability documentation after this change
- **THEN** they can follow steps to locate a checkout TraceId and interpret app, DB/Redis, and async spans without expecting mesh proxy spans in the default view

### Requirement: HTTP and messaging spans use operation-centric names

Marketplace tracing SHALL name user-facing entry spans after the operation (HTTP method + route pattern, gRPC method, or messaging process topic), not after the process or mesh proxy identity. `service.name` remains the logical service identity (`web`, `orders`, …).

#### Scenario: Explore lists readable operations for web

- **WHEN** an operator searches recent traces for service `web` after this change
- **THEN** server span names look like `GET /products` or `POST /checkout` (route patterns), not the bare string `web`

#### Scenario: gRPC and Kafka entry spans remain the deep-dive parents

- **WHEN** an instrumented gRPC handler or Kafka consumer runs with an active parent context
- **THEN** DB and Redis child spans nest under the existing `otelgrpc` / messaging process span without requiring an extra hand-rolled domain span in each service method

### Requirement: Shared Postgres opener emits spans for all queries

`shared/runtime.OpenPostgres` (or the shared path all marketplace services use to obtain `*sql.DB`) SHALL return a database handle instrumented so that queries and transactions executed with a request context export OpenTelemetry spans to VictoriaTraces.

#### Scenario: sqlc query under a gRPC span is visible

- **WHEN** a domain service runs a sqlc query during an instrumented gRPC request
- **THEN** VictoriaTraces shows one or more DB child spans under that TraceId for the query work

#### Scenario: Statement attributes omit bound secrets

- **WHEN** a traced query span is exported
- **THEN** span attributes may include a truncated SQL statement but MUST NOT include bound parameter values that could contain secrets

### Requirement: Shared Redis opener emits spans for commands

`shared/runtime.OpenRedis` SHALL enable OpenTelemetry instrumentation for the go-redis client so cart Redis commands executed with a request context appear as child spans in VictoriaTraces.

#### Scenario: Cart Redis work is visible under the request TraceId

- **WHEN** cart handles a traced gRPC call that reads or writes Redis
- **THEN** Redis command spans appear under the same TraceId as the cart server span
