## ADDED Requirements

### Requirement: Web exports traces and injects gRPC context

The web service SHALL export OpenTelemetry spans to VictoriaTraces and inject W3C trace context on outgoing gRPC calls used for browser and hosted-payment callback flows so downstream services continue the same TraceId.

#### Scenario: Outgoing gRPC calls carry traceparent

- **WHEN** web invokes an internal gRPC API while handling a traced request
- **THEN** the outgoing client call includes W3C trace context derived from the active span

#### Scenario: Hosted payment callback is traced

- **WHEN** web handles `POST /callbacks/hosted-payment`
- **THEN** the request is traced and downstream payment gRPC work continues the same TraceId when instrumentation is enabled
