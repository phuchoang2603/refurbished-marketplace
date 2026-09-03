## ADDED Requirements

### Requirement: VictoriaMetrics accepts application OTLP metrics

The platform observability stack SHALL remain the destination for application OpenTelemetry metrics visualized in Grafana. Marketplace services SHALL push OTLP metrics to VMSingle. An OpenTelemetry Collector is not required for this path.

#### Scenario: Grafana uses VictoriaMetrics for app RED

- **WHEN** operators inspect application request/error/duration after metrics export is enabled
- **THEN** they use the existing Grafana VictoriaMetrics datasource rather than Hubble or Istio scrapes

### Requirement: Marketplace RED dashboard is provisioned

The observability chart SHALL provision a repository-owned Grafana dashboard for marketplace HTTP/gRPC RED from application OpenTelemetry metrics.

#### Scenario: Custom marketplace dashboards load

- **WHEN** the observability chart is synced with custom dashboards enabled
- **THEN** Grafana includes the Marketplace RED dashboard in the Marketplace folder

## MODIFIED Requirements

### Requirement: Backend-first scope

The observability stack SHALL provide metrics, logs, and traces backends. Custom per-service `/metrics` endpoints remain out of scope. Marketplace services SHALL emit structured JSON logs to stdout for VLAgent collection when structured logging is enabled, MAY emit OTLP traces into VictoriaTraces when distributed tracing is enabled, and SHALL emit OTLP metrics into VictoriaMetrics when application RED metrics are enabled.

#### Scenario: Service Prometheus endpoints remain unused for RED

- **WHEN** the platform observability stack is deployed
- **THEN** no Go service is required to add a new `/metrics` endpoint for RED or for platform stack closure

#### Scenario: Application structured logs use existing VL pipeline

- **WHEN** marketplace services emit JSON slog lines to stdout
- **THEN** VLAgent continues to collect those lines into VictoriaLogs without requiring a separate application log exporter

#### Scenario: Application traces use VictoriaTraces

- **WHEN** distributed tracing is enabled for marketplace workloads
- **THEN** Go services MAY export OTLP spans to VictoriaTraces for Grafana Explore

#### Scenario: Application metrics use VictoriaMetrics

- **WHEN** application RED metrics are enabled for marketplace workloads
- **THEN** Go services export OTLP metrics to VictoriaMetrics for Grafana dashboards

### Requirement: Observability documentation

The repository SHALL document how developers and operators access Grafana, verify scrape health, use Trace → logs correlation for marketplace TraceIds, and use application RED metrics in Grafana.

#### Scenario: Developer opens Grafana

- **WHEN** observability is deployed
- **THEN** documentation explains the Grafana public hostname or port-forward and basic login/access path

#### Scenario: Operator verifies scrape health

- **WHEN** observability is deployed
- **THEN** documentation explains how to verify that scrape targets are healthy

#### Scenario: Operator correlates traces to logs

- **WHEN** structured application logging is enabled
- **THEN** documentation explains how to filter VictoriaLogs by `service` and `trace_id` and how to use Grafana Trace → logs

#### Scenario: Operator views application RED

- **WHEN** application OTEL metrics are enabled
- **THEN** documentation explains the metrics OTLP destination, the Marketplace RED dashboard, and that Hubble is not the RED path
