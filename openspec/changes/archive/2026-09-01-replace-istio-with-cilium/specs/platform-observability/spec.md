## MODIFIED Requirements

### Requirement: Victoria observability stack

The repository SHALL provide a Helm wrapper chart for deploying the VictoriaMetrics Kubernetes metrics, logs, traces, dashboards, and alerting stack. Chart defaults SHALL be the full platform profile (node-exporter, kube-state-metrics, Alertmanager, default dashboards, staging-class PVC sizes). A Colima apps-only default overlay SHALL NOT exist.

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
- **THEN** VMSingle requests `20Gi`, VLSingle requests `20Gi`, and VTSingle requests `10Gi` of storage (no separate 5Gi/5Gi/2Gi Colima profile)

#### Scenario: Stack includes logs backend

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaLogs single-node storage and VLAgent collection according to chart values

#### Scenario: Stack includes traces backend

- **WHEN** the observability chart is rendered
- **THEN** it includes VictoriaTraces single-node storage and a Grafana VictoriaTraces datasource according to chart values

### Requirement: Local Argo deploys observability

Argo CD on Talos SHALL deploy the observability stack into the `monitoring` namespace using those full-platform chart defaults (not `local-root` apps-only values).

#### Scenario: Argo includes observability stack

- **WHEN** the Talos app-of-apps syncs
- **THEN** Argo CD manages an observability Application that deploys into `monitoring`

#### Scenario: Grafana is reachable via Gateway or port-forward

- **WHEN** the observability stack is healthy
- **THEN** a Cilium Gateway/HTTPRoute exposes Grafana (dev: `grafana-dev.phuchoang.sbs`) and documentation also explains how to port-forward Grafana in the `monitoring` namespace

### Requirement: Backend-first scope

The observability stack SHALL provide metrics, logs, and traces backends. Custom per-service `/metrics` endpoints remain out of scope for platform closure. Marketplace services SHALL emit structured JSON logs to stdout for VLAgent collection when structured logging is enabled, and MAY emit OTLP traces into VictoriaTraces when distributed tracing is enabled.

#### Scenario: Service instrumentation is deferred

- **WHEN** the platform observability stack is deployed
- **THEN** no Go service is required to add a new `/metrics` endpoint for platform stack closure

#### Scenario: Application structured logs use existing VL pipeline

- **WHEN** marketplace services emit JSON slog lines to stdout
- **THEN** VLAgent continues to collect those lines into VictoriaLogs without requiring a separate application log exporter

#### Scenario: Application traces use VictoriaTraces

- **WHEN** distributed tracing is enabled for marketplace workloads
- **THEN** Go services MAY export OTLP spans to VictoriaTraces for Grafana Explore

### Requirement: VictoriaTraces accepts application OTLP

The platform observability stack SHALL remain the destination for distributed traces visualized in Grafana, including spans exported by marketplace services. Mesh, Hubble, or Gateway proxy tracing is not required.

#### Scenario: Grafana still uses VictoriaTraces

- **WHEN** operators inspect traces after application exporters are enabled
- **THEN** they use the existing Grafana VictoriaTraces datasource rather than a temporary tracing UI

## REMOVED Requirements

### Requirement: Istio L7 metrics scrapes target waypoint and ingress only

The observability chart SHALL scrape Istio L7 metrics from the marketplace waypoint and ingress Gateway proxies and SHALL NOT scrape istiod, ztunnel, or istio-cni as part of the default Istio scrape set.

#### Scenario: Waypoint and ingress are scraped

- **WHEN** `istioScrapes` is enabled in the observability chart
- **THEN** VMPodScrape (or equivalent) targets exist for the ecommerce waypoint and the ecommerce ingress Gateway proxies

#### Scenario: Control-plane ambient scrapes are absent

- **WHEN** `istioScrapes` is enabled in the observability chart
- **THEN** the chart does not create scrape targets for istiod, ztunnel, or istio-cni
