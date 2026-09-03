# App OTEL Metrics

## Purpose

Define application-layer HTTP and gRPC request/error/duration metrics exported from marketplace services via OpenTelemetry into VictoriaMetrics, visualized in Grafana, without Hubble, Istio scrapes, or per-service Prometheus `/metrics` endpoints.

## Requirements

### Requirement: Marketplace services export HTTP and gRPC RED metrics

Instrumented marketplace services SHALL export request rate, error, and duration metrics for inbound HTTP (web) and inbound and outbound gRPC using OpenTelemetry metrics. Metrics SHALL be sent to the platform VictoriaMetrics backend over OTLP. Spans SHALL continue to use the existing VictoriaTraces OTLP path. Metrics SHALL NOT be sent to VictoriaTraces.

#### Scenario: Web HTTP RED is stored in VictoriaMetrics

- **WHEN** `web` handles browser or callback HTTP traffic with metrics export enabled
- **THEN** VictoriaMetrics receives HTTP server request duration (and related RED) series labeled by service, method, status, and low-cardinality route pattern rather than raw URL paths

#### Scenario: gRPC RED is stored in VictoriaMetrics

- **WHEN** an instrumented gRPC server or client handles a marketplace RPC with metrics export enabled
- **THEN** VictoriaMetrics receives RPC duration (and related RED) series labeled by service, RPC method, and status

#### Scenario: Metrics and traces use different backends

- **WHEN** a marketplace service exports both traces and metrics
- **THEN** spans arrive at VictoriaTraces and metrics arrive at VictoriaMetrics

#### Scenario: Empty metrics endpoint disables export

- **WHEN** the metrics OTLP endpoint is unset
- **THEN** the service still starts and does not require VictoriaMetrics to be reachable for process startup

### Requirement: Grafana Marketplace RED dashboard

The observability stack SHALL deploy a Grafana dashboard that shows marketplace HTTP and gRPC request rate, error ratio, and latency from application OpenTelemetry metrics. The dashboard SHALL NOT use Hubble series (`hubble_http_*`), Istio series (`istio_requests_total`), or Gateway proxy metrics as its primary source.

#### Scenario: Dashboard is present after sync

- **WHEN** Grafana marketplace dashboards are synced after this change
- **THEN** a Marketplace RED dashboard is available that queries VictoriaMetrics application metrics

#### Scenario: Operator views service SLIs without Hubble

- **WHEN** an operator opens the Marketplace RED dashboard
- **THEN** they can see request rate, error ratio, and latency for `web` HTTP and internal gRPC services without enabling Hubble

### Requirement: No Prometheus scrape path for app RED

Marketplace services SHALL NOT expose a new `/metrics` HTTP endpoint as the RED export path. RED SHALL use OTLP push to VictoriaMetrics.

#### Scenario: No new app scrape targets required

- **WHEN** application RED is enabled
- **THEN** closure does not depend on VMPodScrape or VMServiceScrape of marketplace application `/metrics` ports
