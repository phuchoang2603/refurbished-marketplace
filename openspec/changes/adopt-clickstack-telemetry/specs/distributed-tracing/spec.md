## RENAMED Requirements

- FROM: `### Requirement: Shared OpenTelemetry bootstrap exports to VictoriaTraces`
- TO: `### Requirement: Shared OpenTelemetry bootstrap exports to the platform agent`

## MODIFIED Requirements

### Requirement: Shared OpenTelemetry bootstrap exports to the platform agent

The repository SHALL provide a shared Go OpenTelemetry bootstrap under `shared/observe/trace` that configures a tracer provider, W3C Trace Context propagation, and OTLP export to the endpoint in the standard `OTEL_EXPORTER_OTLP_*` environment variables. The chart SHALL point that endpoint at the platform `otel-agent`. The bootstrap SHALL NOT contain backend-specific endpoint defaults or paths.

#### Scenario: Service starts with tracing configured

- **WHEN** a marketplace service starts with `OTEL_EXPORTER_OTLP_ENDPOINT` set to the platform agent
- **THEN** spans created by that service reach the ClickHouse store and are visible in HyperDX

#### Scenario: Service starts without an endpoint

- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is empty
- **THEN** the service starts with a no-op exporter and still propagates W3C context

#### Scenario: W3C is the propagation format

- **WHEN** the shared tracing bootstrap configures propagators
- **THEN** it uses W3C `traceparent` / `tracestate` so TraceIds continue across HTTP, gRPC, and Kafka hops between marketplace services

### Requirement: Consumers continue traces as child spans

Kafka consumers for marketplace domain events SHALL extract W3C context from message headers and create child spans of the upstream context (parent–child), not span-links-only, for the v1 visualization model.

#### Scenario: Inventory handles orders.created under parent context

- **WHEN** the inventory consumer processes `orders.created` with a `traceparent` header
- **THEN** it creates a child span under that TraceId visible in HyperDX

#### Scenario: Payment outbox consumer path continues context

- **WHEN** a consumer processes a payment outbox–routed event that carries `traceparent`
- **THEN** it creates a child span under the same TraceId

### Requirement: End-to-end checkout TraceId is verifiable

A Talos-dev checkout and hosted-payment callback SHALL produce a single connected TraceId spanning web, domain services, outbox/Debezium, and consumers as documented. Kafka Connect SHALL export its spans to the same platform agent. Mesh proxy spans are not required for verification.

#### Scenario: Checkout waterfall is connected

- **WHEN** an operator places an order through Talos-dev checkout
- **THEN** HyperDX shows one TraceId covering web → CreateOrder → outbox → Debezium → inventory handling, including DB child spans where those services query Postgres

#### Scenario: Hosted payment callback is connected

- **WHEN** an operator completes a hosted-payment success or failure callback on Talos-dev
- **THEN** HyperDX shows one TraceId covering the callback → payment gRPC → payment outbox path as applicable

#### Scenario: Mesh proxy services are absent from the waterfall

- **WHEN** an operator opens a checkout TraceId after mesh tracing resources are removed
- **THEN** the waterfall does not include `ecommerce-ingress` or `ecommerce-waypoint` spans

### Requirement: Tracing documentation

The repository SHALL document the end-to-end tracing architecture, TraceId joining rules for application and async hops, outbox/Debezium configuration, operation-centric naming expectations, DB/Redis child spans, and HyperDX verification steps that do not depend on Gateway or Hubble proxy spans.

#### Scenario: Contributor finds the tracing guide

- **WHEN** a contributor opens observability documentation after this change
- **THEN** they can follow steps to locate a checkout TraceId in HyperDX and interpret app, DB/Redis, and async spans without expecting mesh or Gateway proxy spans

### Requirement: Shared Postgres opener emits spans for all queries

`shared/runtime.OpenPostgres` (or the shared path all marketplace services use to obtain `*sql.DB`) SHALL return a database handle instrumented so that queries and transactions executed with a request context export OpenTelemetry spans to the platform agent.

#### Scenario: sqlc query under a gRPC span is visible

- **WHEN** a domain service runs a sqlc query during an instrumented gRPC request
- **THEN** HyperDX shows one or more DB child spans under that TraceId for the query work

#### Scenario: Statement attributes omit bound secrets

- **WHEN** a traced query span is exported
- **THEN** span attributes may include a truncated SQL statement but MUST NOT include bound parameter values that could contain secrets

### Requirement: Shared Redis opener emits spans for commands

`shared/runtime.OpenRedis` SHALL enable OpenTelemetry instrumentation for the go-redis client so cart Redis commands executed with a request context appear as child spans in HyperDX.

#### Scenario: Cart Redis work is visible under the request TraceId

- **WHEN** cart handles a traced gRPC call that reads or writes Redis
- **THEN** Redis command spans appear under the same TraceId as the cart server span
