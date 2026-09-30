# Platform Observability

## Purpose

Define how marketplace services consume the telemetry pipeline provided by `talos-proxmox`: OTLP traces and metrics to the node-local `otel-agent`, stdout JSON logs collected by that agent, and the prod ClickHouse store viewed in HyperDX.

## Requirements

### Requirement: Platform observability consumption contract

Marketplace SHALL export OTLP traces and metrics directly to `otel-agent.observability.svc.cluster.local:4317` and write structured stdout logs for platform collection. Marketplace SHALL NOT expose a metrics scrape port or render scrape discovery resources. Deployment guidance SHALL identify a ready `otel-agent` in the target cluster, and the prod ClickHouse store and HyperDX, as prerequisites.

#### Scenario: Application telemetry remains connected

- **WHEN** marketplace deploys against the prepared platform
- **THEN** its metrics, logs, and traces reach the platform store through `otel-agent` without a duplicate telemetry stack

#### Scenario: Store is unreachable

- **WHEN** the prod store or ingest gateway is unavailable
- **THEN** marketplace services keep serving traffic, and only telemetry delivery is affected

### Requirement: Platform agent accepts application OTLP

Marketplace workloads SHALL export traces and metrics over OTLP/gRPC to the platform OpenTelemetry agent at `otel-agent.observability.svc.cluster.local:4317` in the cluster they run in. Marketplace SHALL NOT address the prod ingest gateway, ClickHouse, or HyperDX directly, and SHALL NOT hold a telemetry ingest credential.

#### Scenario: Workload exports telemetry

- **WHEN** a marketplace service on prod exports a span or metric
- **THEN** it sends OTLP to the prod `otel-agent` Service without credentials, and the telemetry appears in HyperDX with `deployment.environment=prod`

#### Scenario: Environment labels come from the platform

- **WHEN** marketplace telemetry is stored
- **THEN** its cluster name and environment are the values set by the platform agent, not values configured in the marketplace chart

### Requirement: Marketplace telemetry carries its service identity

Every marketplace trace, metric, and stdout log record stored by the platform SHALL carry `service.name` equal to the marketplace service name (`web`, `orders`, `payment`, …), so HyperDX can filter all three signals by service.

#### Scenario: stdout logs have a service name

- **WHEN** the platform agent collects a marketplace container's stdout log line
- **THEN** the stored record's `ServiceName` is that service's name rather than empty or a pod name

#### Scenario: Pushed signals have a service name

- **WHEN** a marketplace service exports spans or metrics
- **THEN** the stored records' `ServiceName` is that service's name

### Requirement: Marketplace does not depend on removed telemetry resources

The marketplace Helm releases SHALL NOT render VictoriaMetrics, VictoriaLogs, or VictoriaTraces custom resources, Grafana dashboard ConfigMaps, or any object in the `monitoring` namespace, and SHALL NOT reference Victoria service addresses.

#### Scenario: Sync on the ClickStack platform

- **WHEN** the marketplace roots sync on a cluster that has `otel-agent` but no Victoria CRDs and no `monitoring` namespace
- **THEN** every marketplace Application reaches Synced without errors about missing kinds or namespaces

### Requirement: Trace to logs correlation in HyperDX

Operators SHALL be able to open a marketplace trace in HyperDX and see the log records that share its TraceId, and open a log record that carries a trace ID and navigate to its trace.

#### Scenario: Operator jumps from span to logs

- **WHEN** an operator opens a marketplace checkout trace in HyperDX
- **THEN** HyperDX shows the marketplace JSON log lines written during that trace

#### Scenario: Operator jumps from log to trace

- **WHEN** an operator opens a marketplace log record that has a trace ID
- **THEN** HyperDX offers navigation to that trace

### Requirement: Marketplace consumes the platform telemetry pipeline

Marketplace SHALL consume the telemetry pipeline provided by talos-proxmox. Marketplace services SHALL emit structured JSON logs to stdout for platform agent collection when structured logging is enabled, SHALL export OTLP traces to the platform agent when distributed tracing is enabled, and SHALL export OTLP metrics to the platform agent when application RED metrics are enabled.

#### Scenario: Application structured logs use the platform log pipeline

- **WHEN** marketplace services emit JSON slog lines to stdout
- **THEN** the platform agent collects those lines into the ClickHouse store without requiring an application log exporter

#### Scenario: Application traces use the platform agent

- **WHEN** distributed tracing is enabled for marketplace workloads
- **THEN** Go services and Kafka Connect export OTLP spans to the platform agent, and operators view them in HyperDX

#### Scenario: Application metrics use the platform agent

- **WHEN** application RED metrics are enabled for marketplace workloads
- **THEN** Go services push OTLP metrics to the platform agent, and operators view them in HyperDX

### Requirement: HyperDX observability documentation

The repository SHALL link to talos-proxmox for stack installation and administration, and SHALL document how developers and operators reach HyperDX, verify that each marketplace signal arrives, navigate from traces to logs, and query application RED metrics in HyperDX.

#### Scenario: Developer opens HyperDX

- **WHEN** observability is deployed
- **THEN** documentation gives the internal HyperDX address and points to the platform guide for account access

#### Scenario: Operator verifies telemetry arrival

- **WHEN** observability is deployed
- **THEN** documentation explains how to confirm that marketplace traces, metrics, and logs arrive for each service and environment

#### Scenario: Operator correlates traces to logs

- **WHEN** structured application logging is enabled
- **THEN** documentation explains how to filter HyperDX logs by service, trace ID, and `order_id`, and how to open a trace's logs

#### Scenario: Operator views application RED

- **WHEN** application OTEL metrics are enabled
- **THEN** documentation gives HyperDX queries for request rate, error ratio, and p95 latency for HTTP and gRPC per service
