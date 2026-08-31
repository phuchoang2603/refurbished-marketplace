## MODIFIED Requirements

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

### Requirement: VictoriaTraces accepts application and mesh OTLP

The platform observability stack SHALL remain the destination for distributed traces visualized in Grafana, including spans exported by marketplace services. Mesh or Gateway proxy tracing is not required.

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
