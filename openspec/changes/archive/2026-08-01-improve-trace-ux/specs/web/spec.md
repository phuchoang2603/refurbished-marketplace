## MODIFIED Requirements

### Requirement: Web owns the public browser edge

The web service MUST own the public browser surface, authorization boundary, browser auth cookies, and browser-facing UI routes, and it MUST organize those routes so public, authenticated, and non-browser concerns can apply middleware consistently while keeping unrelated browser routes available when an individual downstream domain service is unavailable.

#### Scenario: A protected route is called

- **WHEN** a client calls a protected browser endpoint
- **THEN** the web service SHALL validate the browser auth cookie and forward trusted identity to internal services

#### Scenario: A browser form is submitted

- **WHEN** a browser submits a web UI form
- **THEN** the web service SHALL process the form at the browser edge and translate successful actions into internal gRPC calls

#### Scenario: A non-browser route is called

- **WHEN** a client calls a health or simulator webhook route
- **THEN** the web service SHALL keep that route outside browser-auth middleware and preserve its documented non-browser contract

#### Scenario: A browser request enters the router

- **WHEN** a browser request enters the web router
- **THEN** the web service SHALL apply request-scoped OpenTelemetry middleware at the web edge so handlers execute with tracing context available on the request, name the HTTP server span using the HTTP method and matched chi route pattern (for example `GET /products/{id}`), set `http.route` to that pattern, and SHALL emit a structured JSON access log for the request instead of chi's text request logger

#### Scenario: A downstream service is unavailable for one feature

- **WHEN** one downstream domain service is unavailable during a browser request
- **THEN** the web service SHALL keep unrelated browser routes and the shared shell available instead of treating the whole browser edge as unavailable

### Requirement: Web exports traces and injects gRPC context

The web service SHALL export OpenTelemetry spans to VictoriaTraces and inject W3C trace context on outgoing gRPC calls used for browser and hosted-payment callback flows so downstream services continue the same TraceId. HTTP server span names SHALL use route patterns rather than the middleware operation string or service name alone.

#### Scenario: Outgoing gRPC calls carry traceparent

- **WHEN** web invokes an internal gRPC API while handling a traced request
- **THEN** the outgoing client call includes W3C trace context derived from the active span

#### Scenario: Hosted payment callback is traced

- **WHEN** web handles `POST /callbacks/hosted-payment`
- **THEN** the request produces a server span named with method and route pattern continuing into downstream payment gRPC work on the same TraceId when instrumentation is enabled

#### Scenario: Span names avoid raw path cardinality

- **WHEN** a traced request matches a parameterized chi route
- **THEN** the HTTP server span name and `http.route` use the route pattern (with placeholders) rather than the raw URL path containing concrete IDs
