# Platform Observability

## Purpose

Define the VictoriaMetrics Kubernetes metrics, logs, traces, dashboards, and alerting stack delivered through a repository-owned Helm wrapper and staging GitOps, without requiring application instrumentation in the first slice.

## Requirements

### Requirement: Victoria observability stack

The repository SHALL provide a local Helm wrapper chart for deploying the VictoriaMetrics Kubernetes metrics, logs, traces, dashboards, and alerting stack.

#### Scenario: Wrapper chart defines observability stack

- **WHEN** the observability chart dependencies are built
- **THEN** the chart includes `victoria-metrics-k8s-stack` from `https://victoriametrics.github.io/helm-charts/` as a dependency pinned to version `0.86.0`

#### Scenario: Stack includes core metrics components

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaMetrics metrics storage, VMAgent scraping, Grafana, Alertmanager, kube-state-metrics, and node-exporter components according to chart values

#### Scenario: Stack uses single-node backends

- **WHEN** the observability chart is rendered for this change
- **THEN** it enables VMSingle, VLSingle, and VTSingle and does not require VMCluster, VLCluster, or VTCluster

#### Scenario: Stack uses default storage class

- **WHEN** persistent storage is configured for VMSingle, VLSingle, or VTSingle
- **THEN** the chart does not override the cluster default storage class

#### Scenario: Stack uses initial retention periods

- **WHEN** single-node backend retention is configured
- **THEN** metrics retention is `7d`, logs retention is `3d`, and traces retention is `3d`

#### Scenario: Stack uses local PVC sizes

- **WHEN** the observability chart is rendered for local validation
- **THEN** VMSingle requests `5Gi`, VLSingle requests `5Gi`, and VTSingle requests `2Gi` of storage

#### Scenario: Stack uses staging PVC sizes

- **WHEN** the observability chart is rendered for staging
- **THEN** VMSingle requests `20Gi`, VLSingle requests `20Gi`, and VTSingle requests `10Gi` of storage

#### Scenario: Stack includes logs backend

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaLogs single-node storage and VLAgent collection according to chart values

#### Scenario: Stack includes traces backend

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaTraces single-node storage and a Grafana VictoriaTraces datasource according to chart values

### Requirement: Local Argo deploys observability

Local Argo CD (`values-local.yaml` on the shared app-of-apps chart) SHALL deploy the observability stack into the `monitoring` namespace using chart default values.

#### Scenario: Local Argo includes observability stack

- **WHEN** the local app-of-apps syncs
- **THEN** Argo CD manages an observability Application that deploys into `monitoring`

#### Scenario: Local Grafana is reachable via port-forward

- **WHEN** the local observability stack is healthy
- **THEN** documentation explains how to port-forward Grafana in the `monitoring` namespace

### Requirement: Backend-first scope

The observability stack SHALL provide metrics, logs, and traces backends without requiring Go services to emit new application metrics, logs, or traces in this change.

#### Scenario: Service instrumentation is deferred

- **WHEN** this change is implemented
- **THEN** no Go service is required to add a new `/metrics` endpoint for closure

#### Scenario: Application log wiring is deferred

- **WHEN** this change is implemented
- **THEN** no Go service is required to change its logging implementation for closure

#### Scenario: Application trace wiring is deferred

- **WHEN** this change is implemented
- **THEN** no Go service is required to emit OTLP traces for closure

#### Scenario: Istio supplies service request metrics later

- **WHEN** Istio observe mode is available
- **THEN** request rate, request latency, and request error ratio dashboards can use Istio L7 metrics instead of per-service custom instrumentation where those metrics are sufficient

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
- **THEN** it has a VictoriaTraces datasource configured through the Jaeger-compatible API

#### Scenario: Alertmanager is available

- **WHEN** the observability stack is running
- **THEN** Alertmanager is deployed and can receive alert rules from the stack configuration

### Requirement: Observability documentation

The repository SHALL document how developers and operators access Grafana and verify scrape health.

#### Scenario: Developer opens Grafana

- **WHEN** observability is deployed
- **THEN** documentation explains the Grafana port-forward and basic login/access path

#### Scenario: Operator verifies scrape health

- **WHEN** staging observability is deployed
- **THEN** documentation explains how to verify that scrape targets are healthy
