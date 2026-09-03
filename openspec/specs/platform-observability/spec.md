# Platform Observability

## Purpose

Define the VictoriaMetrics Kubernetes metrics, logs, traces, dashboards, and alerting stack delivered through a repository-owned Helm wrapper and Talos GitOps, without requiring per-service `/metrics` endpoints for platform closure.

## Requirements

### Requirement: Victoria observability stack

The repository SHALL provide a Helm wrapper chart for deploying the VictoriaMetrics Kubernetes metrics, logs, traces, dashboards, and alerting stack. Chart defaults SHALL be the full platform profile (node-exporter, kube-state-metrics, Alertmanager, default dashboards, Talos PVC sizes).

#### Scenario: Wrapper chart defines observability stack

- **WHEN** the observability chart dependencies are built
- **THEN** the chart includes `victoria-metrics-k8s-stack` from `https://victoriametrics.github.io/helm-charts/` as a dependency pinned to version `0.86.0`

#### Scenario: Stack includes core metrics components

- **WHEN** the observability chart is rendered with default values
- **THEN** it includes VictoriaMetrics metrics storage, VMAgent scraping, Grafana, Alertmanager, kube-state-metrics, and node-exporter

#### Scenario: Stack uses single-node backends

- **WHEN** the observability chart is rendered for this change
- **THEN** it enables VMSingle, VLSingle, and VTSingle and does not require VMCluster, VLCluster, or VTCluster

#### Scenario: Stack uses default storage class

- **WHEN** persistent storage is configured for VMSingle, VLSingle, or VTSingle
- **THEN** the chart does not override the cluster default storage class

#### Scenario: Stack uses initial retention periods

- **WHEN** single-node backend retention is configured
- **THEN** metrics retention is `7d`, logs retention is `3d`, and traces retention is `3d`

#### Scenario: Stack uses one PVC size profile

- **WHEN** the observability chart is rendered
- **THEN** VMSingle requests `20Gi`, VLSingle requests `20Gi`, and VTSingle requests `10Gi` of storage

#### Scenario: Stack includes logs backend

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaLogs single-node storage and VLAgent collection according to chart values

#### Scenario: Stack includes traces backend

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaTraces single-node storage and a Grafana VictoriaTraces datasource according to chart values

### Requirement: Argo deploys observability

Argo CD on Talos SHALL deploy the observability stack into the `monitoring` namespace using those full-platform chart defaults.

#### Scenario: Argo includes observability stack

- **WHEN** the Talos app-of-apps syncs
- **THEN** Argo CD manages an observability Application that deploys into `monitoring`

#### Scenario: Grafana is reachable via Gateway or port-forward

- **WHEN** the observability stack is healthy
- **THEN** a Cilium Gateway/HTTPRoute exposes Grafana (dev: `grafana-dev.phuchoang.sbs`) and documentation also explains how to port-forward Grafana in the `monitoring` namespace

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

### Requirement: Grafana datasources and alerting baseline

The observability stack SHALL include Grafana datasources and an initial alerting path suitable for platform and future service dashboards.

#### Scenario: Grafana has metrics datasource

- **WHEN** Grafana starts from the observability stack
- **THEN** it has a VictoriaMetrics-compatible metrics datasource configured

#### Scenario: Grafana has logs datasource

- **WHEN** Grafana starts from the observability stack
- **THEN** it has a VictoriaLogs datasource configured with the required Grafana plugin

#### Scenario: Grafana has traces datasource

- **WHEN** Grafana starts from the observability stack
- **THEN** it has a VictoriaTraces datasource configured through the Tempo-compatible API (`/select/tempo`)

#### Scenario: Alertmanager is available

- **WHEN** the observability stack is running
- **THEN** Alertmanager is deployed and can receive alert rules from the stack configuration

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

### Requirement: VictoriaTraces accepts application OTLP

The platform observability stack SHALL remain the destination for distributed traces visualized in Grafana, including spans exported by marketplace services. Mesh, Hubble, or Gateway proxy tracing is not required.

#### Scenario: Grafana still uses VictoriaTraces

- **WHEN** operators inspect traces after application exporters are enabled
- **THEN** they use the existing Grafana VictoriaTraces datasource rather than a temporary tracing UI

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

### Requirement: Trace to logs correlation in Grafana

The observability stack SHALL configure the Grafana VictoriaTraces (Tempo) datasource so operators can navigate from a span to VictoriaLogs lines filtered by TraceId (`trace_id`).

#### Scenario: Tempo datasource links to VictoriaLogs

- **WHEN** Grafana starts from the observability chart after this change
- **THEN** the VictoriaTraces Tempo datasource includes Trace → logs configuration targeting the VictoriaLogs datasource using the log field `trace_id`

#### Scenario: Operator jumps from span to logs

- **WHEN** an operator opens a marketplace span in Grafana Explore or Traces Drilldown and uses Trace → logs
- **THEN** Grafana shows VictoriaLogs results for that TraceId when matching JSON log lines exist
