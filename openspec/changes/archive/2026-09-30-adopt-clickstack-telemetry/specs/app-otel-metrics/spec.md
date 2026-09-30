## ADDED Requirements

### Requirement: Marketplace services push HTTP and gRPC RED metrics

Instrumented marketplace services SHALL export request rate, error, and duration metrics for inbound HTTP (web) and inbound and outbound gRPC using OpenTelemetry metrics. Metrics SHALL keep OpenTelemetry semantic-convention names and attributes, and SHALL carry `service.name` as a resource attribute. Metrics and spans SHALL use the same platform agent endpoint.

#### Scenario: Web HTTP RED is stored

- **WHEN** `web` handles browser or callback HTTP traffic
- **THEN** the store receives `http.server.request.duration` data points for service `web` with method, status, and low-cardinality route pattern attributes, not raw URL paths

#### Scenario: gRPC RED is stored

- **WHEN** an instrumented gRPC server or client handles a marketplace RPC
- **THEN** the store receives RPC duration data points for that service with RPC method and status attributes

#### Scenario: Metrics and traces share the agent

- **WHEN** a marketplace service exports both traces and metrics
- **THEN** both arrive at the platform agent over OTLP and can be filtered by the same `service.name` in HyperDX

### Requirement: OTLP push path for app RED

Marketplace services SHALL push RED metrics over OTLP to the endpoint in `OTEL_EXPORTER_OTLP_ENDPOINT`, on a fixed export interval. Services SHALL NOT open a metrics scrape listener or expose a metrics port.

#### Scenario: Metrics arrive without scraping

- **WHEN** a marketplace service runs on a cluster with the platform agent
- **THEN** its RED metrics reach the ClickHouse store without any scrape configuration, scrape network policy, or metrics Service port

#### Scenario: Unset endpoint still starts

- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is empty
- **THEN** the service still starts, records metrics in-process, and exports nothing

#### Scenario: Metrics are flushed on shutdown

- **WHEN** a marketplace service shuts down gracefully
- **THEN** it exports pending metrics before exiting

## REMOVED Requirements

### Requirement: Marketplace services export HTTP and gRPC RED metrics

**Reason**: It tied RED metrics to a Prometheus scrape into VictoriaMetrics, which no longer exists.
**Migration**: See "Marketplace services push HTTP and gRPC RED metrics".

### Requirement: Grafana Marketplace RED dashboard

**Reason**: Grafana was removed from the platform, and HyperDX dashboards are not managed from Git.
**Migration**: Use the documented HyperDX RED queries in `docs/deployment/observability.md`.

### Requirement: Prometheus scrape path for app RED

**Reason**: VMAgent was removed. The platform's annotation scrape would mislabel the service identity, so OTLP push replaces scraping.
**Migration**: See "OTLP push path for app RED". `METRICS_ADDR`, the `:9100` port, and the `metrics`/`metricsPort` chart keys are removed.
