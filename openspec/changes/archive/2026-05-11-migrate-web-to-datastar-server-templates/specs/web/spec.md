## MODIFIED Requirements

### Requirement: Web owns the public browser edge

The web service MUST own the public browser surface and authorization boundary.

#### Scenario: A protected route is called

- **WHEN** a client calls a protected browser endpoint
- **THEN** the web service SHALL validate the access token and forward trusted identity to internal services

### Requirement: Web exposes the documented public routes

The web service MUST expose the documented browser routes for health, auth, and catalog browsing using server-rendered HTML responses.

#### Scenario: Public route is requested

- **WHEN** a client calls a documented public route
- **THEN** the route SHALL be served as HTML or an HTML fragment without requiring internal service access
