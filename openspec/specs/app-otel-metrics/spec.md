# App OTEL Metrics

## Purpose

Define application-layer HTTP and gRPC request/error/duration metrics from marketplace OpenTelemetry instrumentation, scraped as Prometheus `/metrics` into VictoriaMetrics and visualized in Grafana, without Hubble or Istio scrapes.

## Requirements

### Requirement: Marketplace services export HTTP and gRPC RED metrics

Instrumented marketplace services SHALL export request rate, error, and duration metrics for inbound HTTP (web) and inbound and outbound gRPC using OpenTelemetry metrics recorded against a Prometheus-backed MeterProvider. Metrics SHALL be scraped by the platform VictoriaMetrics agent from `/metrics`. Spans SHALL continue to use the existing VictoriaTraces OTLP path. Metrics SHALL NOT be sent to VictoriaTraces.

#### Scenario: Web HTTP RED is stored in VictoriaMetrics

- **WHEN** `web` handles browser or callback HTTP traffic with the metrics listener enabled
- **THEN** VictoriaMetrics receives HTTP server request duration (and related RED) series labeled by job, method, status, and low-cardinality route pattern rather than raw URL paths

#### Scenario: gRPC RED is stored in VictoriaMetrics

- **WHEN** an instrumented gRPC server or client handles a marketplace RPC with the metrics listener enabled
- **THEN** VictoriaMetrics receives RPC duration (and related RED) series labeled by job, RPC method, and status

#### Scenario: Metrics and traces use different backends

- **WHEN** a marketplace service exports both traces and metrics
- **THEN** spans arrive at VictoriaTraces over OTLP and metrics arrive at VictoriaMetrics via scrape

#### Scenario: Disabled metrics listener still starts

- **WHEN** `METRICS_ADDR` is set to `-`
- **THEN** the service still starts and does not require a scrape listener

### Requirement: Grafana Marketplace RED dashboard

The observability stack SHALL deploy a Grafana dashboard that shows marketplace HTTP and gRPC request rate, error ratio, and latency from scraped application metrics. The dashboard SHALL NOT use Hubble series (`hubble_http_*`), Istio series (`istio_requests_total`), or Gateway proxy metrics as its primary source.

#### Scenario: Dashboard is present after sync

- **WHEN** Grafana marketplace dashboards are synced after this change
- **THEN** a Marketplace RED dashboard is available that queries VictoriaMetrics application metrics

#### Scenario: Operator views service SLIs without Hubble

- **WHEN** an operator opens the Marketplace RED dashboard
- **THEN** they can see request rate, error ratio, and latency for `web` HTTP and internal gRPC services without enabling Hubble

### Requirement: Prometheus scrape path for app RED

Marketplace services SHALL expose `/metrics` as the RED export path. RED SHALL use VMAgent scrape (VMPodScrape) into VictoriaMetrics, not OTLP push.

#### Scenario: App scrape targets are required

- **WHEN** application RED is enabled
- **THEN** closure depends on VMAgent scraping marketplace application `/metrics` ports
