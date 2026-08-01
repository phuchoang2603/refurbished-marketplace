## ADDED Requirements

_(none)_

## REMOVED Requirements

### Requirement: Mesh tracing exports OpenTelemetry spans to VictoriaTraces

**Reason:** Mesh proxy spans (`ecommerce-ingress`, `ecommerce-waypoint`) dominated Explore roots without helping app/DB deep-dives. Application OTEL already joins TraceIds across services; canary readiness later depends on versioned metrics, not Envoy spans.

**Migration:** Remove marketplace Gateway Telemetry (`ecommerce-tracing`), the `mesh.tracing` chart values, the `vtsingle-otlp-plain` DestinationRule, and the istiod `otel-vt` extension provider. Keep ambient enrollment and Istio L7 metrics scrapes.

## MODIFIED Requirements

### Requirement: Mesh telemetry visibility

The system SHALL provide observable Istio telemetry for marketplace service-to-service traffic in staging via L7 **metrics** (and dashboards). Distributed tracing from ingress/waypoint proxies to VictoriaTraces is not part of the marketplace observe path.

#### Scenario: Internal traffic appears in telemetry

- **WHEN** a user exercises the primary marketplace flows in staging
- **THEN** mesh **metrics** show traffic involving `web`, `users`, `products`, `orders`, `cart`, `payment`, and `payment-gateway-simulator` where applicable

#### Scenario: gRPC traffic is distinguishable

- **WHEN** the web service calls internal gRPC services during staging verification
- **THEN** telemetry distinguishes gRPC service calls from opaque TCP traffic where Istio protocol support allows it

#### Scenario: Grafana and VictoriaMetrics are the mesh metrics visualization path

- **WHEN** mesh dashboard verification is documented
- **THEN** the documentation targets Grafana with Istio L7 metrics (e.g. Marketplace Istio RED) rather than requiring proxy spans in VictoriaTraces

#### Scenario: Mesh tracing resources are absent

- **WHEN** the marketplace and istiod charts are deployed after this change
- **THEN** they do not render a tracing Telemetry resource, VictoriaTraces OTEL extension provider, or mesh-tracing DestinationRule for ecommerce ingress/waypoint span export

#### Scenario: Telemetry verification uses platform observability

- **WHEN** Istio metrics and dashboard verification runs in staging
- **THEN** it uses the deployed `platform-observability` stack (`monitoring` namespace, Grafana / VictoriaMetrics) rather than a temporary tracing UI
