## MODIFIED Requirements

### Requirement: Web exports traces and injects gRPC context

The web service SHALL export OpenTelemetry spans to the platform agent and inject W3C trace context on outgoing gRPC calls used for browser and hosted-payment callback flows so downstream services continue the same TraceId. HTTP server span names SHALL use route patterns rather than the middleware operation string or service name alone.

#### Scenario: Outgoing gRPC calls carry traceparent

- **WHEN** web invokes an internal gRPC API while handling a traced request
- **THEN** the outgoing client call includes W3C trace context derived from the active span

#### Scenario: Hosted payment callback is traced

- **WHEN** web handles `POST /callbacks/hosted-payment`
- **THEN** the request produces a server span named with method and route pattern continuing into downstream payment gRPC work on the same TraceId when instrumentation is enabled

#### Scenario: Span names avoid raw path cardinality

- **WHEN** a traced request matches a parameterized chi route
- **THEN** the HTTP server span name and `http.route` use the route pattern (with placeholders) rather than the raw URL path containing concrete IDs
