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
- **THEN** the web service SHALL apply request-scoped OpenTelemetry middleware at the web edge so handlers execute with tracing context available on the request

#### Scenario: A downstream service is unavailable for one feature

- **WHEN** one downstream domain service is unavailable during a browser request
- **THEN** the web service SHALL keep unrelated browser routes and the shared shell available instead of treating the whole browser edge as unavailable
