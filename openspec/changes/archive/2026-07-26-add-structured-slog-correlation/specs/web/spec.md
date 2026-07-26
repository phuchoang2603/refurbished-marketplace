## ADDED Requirements

### Requirement: Web emits structured HTTP access logs

The web service SHALL emit JSON slog access logs for browser and non-browser HTTP requests including method, path, status, duration, and `trace_id` when an OpenTelemetry span is active on the request. The web router SHALL NOT register chi `middleware.RequestID` or chi `middleware.Logger`.

#### Scenario: Request completes with slog access log

- **WHEN** an HTTP request completes through the web router
- **THEN** a JSON access log line includes method, path, status, duration, and `trace_id` when tracing context is present

#### Scenario: Chi RequestID and text Logger are absent

- **WHEN** the web router middleware stack is configured
- **THEN** it does not include `middleware.RequestID` or `middleware.Logger`

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
- **THEN** the web service SHALL apply request-scoped OpenTelemetry middleware at the web edge so handlers execute with tracing context available on the request, and SHALL emit a structured JSON access log for the request instead of chi's text request logger

#### Scenario: A downstream service is unavailable for one feature

- **WHEN** one downstream domain service is unavailable during a browser request
- **THEN** the web service SHALL keep unrelated browser routes and the shared shell available instead of treating the whole browser edge as unavailable
