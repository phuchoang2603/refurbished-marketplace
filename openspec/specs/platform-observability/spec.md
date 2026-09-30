# Platform Observability

## Purpose

Define how marketplace services consume the telemetry pipeline provided by `talos-proxmox`: OTLP traces and metrics to the node-local `otel-agent`, stdout JSON logs collected by that agent, and the prod ClickHouse store viewed in HyperDX.

## Requirements

### Requirement: Backend-first scope

Marketplace SHALL consume the metrics, logs, and traces backends provided by talos-proxmox. Marketplace services SHALL emit structured JSON logs to stdout for VLAgent collection when structured logging is enabled, MAY emit OTLP traces into VictoriaTraces when distributed tracing is enabled, and SHALL expose Prometheus `/metrics` for VMAgent scrape when application RED metrics are enabled.

#### Scenario: Application structured logs use existing VL pipeline

- **WHEN** marketplace services emit JSON slog lines to stdout
- **THEN** VLAgent continues to collect those lines into VictoriaLogs without requiring a separate application log exporter

#### Scenario: Application traces use VictoriaTraces

- **WHEN** distributed tracing is enabled for marketplace workloads
- **THEN** Go services MAY export OTLP spans to VictoriaTraces for Grafana Explore

#### Scenario: Application metrics use VictoriaMetrics scrape

- **WHEN** application RED metrics are enabled for marketplace workloads
- **THEN** Go services expose `/metrics` and VMAgent scrapes them into VictoriaMetrics for Grafana dashboards

### Requirement: Observability documentation

The repository SHALL link to talos-proxmox for stack installation and administration, and SHALL document how developers and operators access Grafana, verify scrape health, use Trace → logs correlation for marketplace TraceIds, and use application RED metrics in Grafana.

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
- **THEN** documentation explains the `/metrics` scrape path, the Marketplace RED dashboard, and the application scrape/export contract

### Requirement: VictoriaTraces accepts application OTLP

The platform observability stack SHALL remain the destination for distributed traces visualized in Grafana, including spans exported by marketplace services. Mesh, Hubble, or Gateway proxy tracing is not required.

#### Scenario: Grafana still uses VictoriaTraces

- **WHEN** operators inspect traces after application exporters are enabled
- **THEN** they use the existing Grafana VictoriaTraces datasource rather than a temporary tracing UI

### Requirement: VictoriaMetrics stores application RED metrics

The platform observability stack SHALL remain the destination for application RED metrics visualized in Grafana. Marketplace services SHALL expose Prometheus `/metrics` for VMAgent scrape into VMSingle. An OpenTelemetry Collector is not required for this path.

#### Scenario: Grafana uses VictoriaMetrics for app RED

- **WHEN** operators inspect application request/error/duration after metrics scrape is enabled
- **THEN** they use the existing Grafana VictoriaMetrics datasource rather than Hubble or Istio scrapes

### Requirement: Marketplace RED dashboard is provisioned

The marketplace release SHALL provision an application-owned Grafana dashboard for marketplace HTTP/gRPC request, error, and duration metrics as a labeled ConfigMap in the existing monitoring namespace. It SHALL use the platform Grafana datasource contract without deploying Grafana or changing platform namespace metadata. Folder organization SHALL follow platform Grafana configuration.

#### Scenario: Custom marketplace dashboards load

- **WHEN** the marketplace release syncs on a ready platform
- **THEN** platform Grafana discovers the ConfigMap and exposes the Marketplace RED dashboard

#### Scenario: Platform remains independently owned

- **WHEN** the marketplace release is removed
- **THEN** only its dashboard ConfigMaps are removed; the monitoring namespace and Grafana remain platform-owned

### Requirement: Marketplace logs dashboard is provisioned

The marketplace release SHALL provision an application-owned Grafana dashboard for marketplace JSON logs with service filtering as a labeled ConfigMap in the existing monitoring namespace. It SHALL use the platform Grafana datasource contract without deploying Grafana or changing platform namespace metadata. Folder organization SHALL follow platform Grafana configuration.

#### Scenario: Custom marketplace logs dashboard loads

- **WHEN** the marketplace release syncs on a ready platform
- **THEN** platform Grafana discovers the ConfigMap and exposes the Marketplace logs dashboard

#### Scenario: Platform remains independently owned

- **WHEN** the marketplace release is removed
- **THEN** only its dashboard ConfigMaps are removed; the monitoring namespace and Grafana remain platform-owned

### Requirement: Trace to logs correlation in Grafana

The marketplace telemetry contract SHALL consume platform configuration of the Grafana VictoriaTraces (Tempo) datasource so operators can navigate from a span to VictoriaLogs using LogsQL on the log field `trace_id`. The link SHALL NOT rely on Loki-style stream selectors.

#### Scenario: Tempo datasource links to VictoriaLogs

- **WHEN** the platform Grafana datasource configuration is verified
- **THEN** the VictoriaTraces Tempo datasource includes Trace → logs configuration targeting the VictoriaLogs datasource with a LogsQL query on `trace_id`

#### Scenario: Operator jumps from span to logs

- **WHEN** an operator opens a marketplace span in Grafana Explore or Traces Drilldown and uses Trace → logs
- **THEN** Grafana shows VictoriaLogs results for that TraceId when matching JSON log lines exist

### Requirement: Log to traces and metrics correlation in Grafana

The marketplace telemetry contract SHALL consume platform-provided VictoriaLogs derived fields so operators can jump from a log line to the matching trace and to application metrics that share the service `job` label.

#### Scenario: Operator jumps from log to trace

- **WHEN** an operator opens a marketplace JSON log that includes `trace_id`
- **THEN** Grafana offers a link to VictoriaTraces for that TraceId

#### Scenario: Operator jumps from log to metrics

- **WHEN** an operator opens a marketplace JSON log that includes `service`
- **THEN** Grafana offers a link to VictoriaMetrics series filtered by that service name

### Requirement: Platform observability consumption contract

Marketplace SHALL retain VMPodScrape discovery of application /metrics on port 9100, the network policy allowing platform VMAgent scrapes, structured stdout logs for platform collection, and direct OTLP trace export to vtsingle-vmks.monitoring.svc.cluster.local:4317. Deployment guidance SHALL identify ready platform datasources, scrape discovery, log collection, and dashboard discovery as prerequisites.

#### Scenario: Application telemetry remains connected

- **WHEN** marketplace deploys against the prepared platform
- **THEN** its metrics, logs, and traces use platform services without a duplicate telemetry stack
