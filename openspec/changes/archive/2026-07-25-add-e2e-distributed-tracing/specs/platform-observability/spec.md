## ADDED Requirements

### Requirement: Istio L7 metrics scrapes target waypoint and ingress only

The observability chart SHALL scrape Istio L7 metrics from the marketplace waypoint and ingress Gateway proxies and SHALL NOT scrape istiod, ztunnel, or istio-cni as part of the default Istio scrape set.

#### Scenario: Waypoint and ingress are scraped

- **WHEN** `istioScrapes` is enabled in the observability chart
- **THEN** VMPodScrape (or equivalent) targets exist for the ecommerce waypoint and the ecommerce ingress Gateway proxies

#### Scenario: Control-plane ambient scrapes are absent

- **WHEN** `istioScrapes` is enabled in the observability chart
- **THEN** the chart does not create scrape targets for istiod, ztunnel, or istio-cni

### Requirement: VictoriaTraces accepts application and mesh OTLP

The platform observability stack SHALL remain the destination for distributed traces visualized in Grafana, including spans exported by marketplace services and Istio OpenTelemetry tracing.

#### Scenario: Grafana still uses VictoriaTraces

- **WHEN** operators inspect traces after application and mesh exporters are enabled
- **THEN** they use the existing Grafana VictoriaTraces datasource rather than a temporary tracing UI

## MODIFIED Requirements

### Requirement: Backend-first scope

The observability stack SHALL provide metrics, logs, and traces backends. Custom per-service `/metrics` endpoints and application log-pipeline changes remain out of scope for platform closure, but marketplace services and Istio MAY emit OTLP traces into VictoriaTraces when distributed tracing is enabled.

#### Scenario: Service instrumentation is deferred

- **WHEN** the platform observability stack is deployed
- **THEN** no Go service is required to add a new `/metrics` endpoint for platform stack closure

#### Scenario: Application log wiring is deferred

- **WHEN** the platform observability stack is deployed
- **THEN** no Go service is required to change its logging implementation for platform stack closure

#### Scenario: Application and mesh traces may use VictoriaTraces

- **WHEN** distributed tracing is enabled for marketplace workloads
- **THEN** Go services and Istio OpenTelemetry tracing MAY export OTLP spans to VictoriaTraces for Grafana Explore

#### Scenario: Istio supplies service request metrics

- **WHEN** Istio L7 metrics are scraped from waypoint and ingress
- **THEN** request rate, request latency, and request error ratio dashboards can use those Istio metrics instead of per-service custom instrumentation where those metrics are sufficient
