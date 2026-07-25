## ADDED Requirements

### Requirement: Mesh tracing exports OpenTelemetry spans to VictoriaTraces

When marketplace Istio L7 proxies are configured for distributed tracing, the system SHALL export OpenTelemetry spans from the ingress Gateway and waypoint to VictoriaTraces using W3C Trace Context so proxy spans can share TraceIds with instrumented applications that propagate `traceparent`.

#### Scenario: Extension provider targets VictoriaTraces

- **WHEN** Istio mesh tracing is enabled for staging
- **THEN** an OpenTelemetry extension provider (or equivalent) is configured to send spans to the platform VictoriaTraces OTLP endpoint

#### Scenario: Waypoint and ingress Telemetry enable tracing

- **WHEN** tracing Telemetry is applied for marketplace edge and east-west L7
- **THEN** the ecommerce ingress Gateway and ecommerce waypoint emit spans for sampled requests

#### Scenario: Joined TraceId requires app header propagation

- **WHEN** an application forwards W3C `traceparent` on outbound calls through the mesh
- **THEN** Istio proxy spans and application spans for that request share the same TraceId in VictoriaTraces

#### Scenario: Ambient L4 is not the distributed tracing surface

- **WHEN** workloads use ambient mode without relying on ztunnel for span trees
- **THEN** distributed mesh tracing is expected from waypoint and ingress L7 proxies rather than ztunnel
