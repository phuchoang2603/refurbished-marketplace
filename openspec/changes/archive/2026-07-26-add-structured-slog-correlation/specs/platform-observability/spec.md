## ADDED Requirements

### Requirement: Trace to logs correlation in Grafana

The observability stack SHALL configure the Grafana VictoriaTraces (Tempo) datasource so operators can navigate from a span to VictoriaLogs lines filtered by TraceId (`trace_id`).

#### Scenario: Tempo datasource links to VictoriaLogs

- **WHEN** Grafana starts from the observability chart after this change
- **THEN** the VictoriaTraces Tempo datasource includes Trace → logs configuration targeting the VictoriaLogs datasource using the log field `trace_id`

#### Scenario: Operator jumps from span to logs

- **WHEN** an operator opens a marketplace span in Grafana Explore or Traces Drilldown and uses Trace → logs
- **THEN** Grafana shows VictoriaLogs results for that TraceId when matching JSON log lines exist

## MODIFIED Requirements

### Requirement: Backend-first scope

The observability stack SHALL provide metrics, logs, and traces backends. Custom per-service `/metrics` endpoints remain out of scope for platform closure. Marketplace services SHALL emit structured JSON logs to stdout for VLAgent collection when structured logging is enabled, and MAY emit OTLP traces into VictoriaTraces when distributed tracing is enabled.

#### Scenario: Service instrumentation is deferred

- **WHEN** the platform observability stack is deployed
- **THEN** no Go service is required to add a new `/metrics` endpoint for platform stack closure

#### Scenario: Application structured logs use existing VL pipeline

- **WHEN** marketplace services emit JSON slog lines to stdout
- **THEN** VLAgent continues to collect those lines into VictoriaLogs without requiring a separate application log exporter

#### Scenario: Application and mesh traces may use VictoriaTraces

- **WHEN** distributed tracing is enabled for marketplace workloads
- **THEN** Go services and Istio OpenTelemetry tracing MAY export OTLP spans to VictoriaTraces for Grafana Explore

#### Scenario: Istio supplies service request metrics

- **WHEN** Istio L7 metrics are scraped from waypoint and ingress
- **THEN** request rate, request latency, and request error ratio dashboards can use those Istio metrics instead of per-service custom instrumentation where those metrics are sufficient

### Requirement: Observability documentation

The repository SHALL document how developers and operators access Grafana, verify scrape health, and use Trace → logs correlation for marketplace TraceIds.

#### Scenario: Developer opens Grafana

- **WHEN** observability is deployed
- **THEN** documentation explains the Grafana port-forward and basic login/access path

#### Scenario: Operator verifies scrape health

- **WHEN** staging observability is deployed
- **THEN** documentation explains how to verify that scrape targets are healthy

#### Scenario: Operator correlates traces to logs

- **WHEN** structured application logging is enabled
- **THEN** documentation explains how to filter VictoriaLogs by `service` and `trace_id` and how to use Grafana Trace → logs
