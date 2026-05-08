## ADDED Requirements

### Requirement: Web owns the REST edge

The web service MUST own the public REST surface and authorization boundary.

#### Scenario: A protected route is called

- **WHEN** a client calls a protected REST endpoint
- **THEN** the web service SHALL validate the access token and forward trusted identity to internal services

### Requirement: Web delegates auth session logic to users

The web service MUST delegate login, refresh, and logout session logic to the users service.

#### Scenario: Login request arrives

- **WHEN** a client calls the login endpoint
- **THEN** the web service SHALL invoke the users service for session issuance

### Requirement: Web exposes the documented public routes

The web service MUST expose the documented REST routes for health, auth, and catalog browsing.

#### Scenario: Public route is requested

- **WHEN** a client calls a documented public route
- **THEN** the route SHALL be served without requiring internal service access
