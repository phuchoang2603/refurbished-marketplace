## MODIFIED Requirements

### Requirement: Web owns the public browser edge

The web service MUST own the public browser surface, authorization boundary, browser auth cookies, and browser-facing UI routes.

#### Scenario: A protected route is called

- **WHEN** a client calls a protected browser endpoint
- **THEN** the web service SHALL validate the browser auth cookie or access token and forward trusted identity to internal services

#### Scenario: A browser form is submitted

- **WHEN** a browser submits a web UI form
- **THEN** the web service SHALL process the form at the browser edge and translate successful actions into internal gRPC calls

### Requirement: Web delegates auth session logic to users

The web service MUST delegate login and logout session logic to the users service while presenting browser-friendly auth UI responses and managing browser cookie persistence.

#### Scenario: Login request arrives

- **WHEN** a client calls the login endpoint
- **THEN** the web service SHALL invoke the users service for session issuance

#### Scenario: Login form is submitted from the browser

- **WHEN** a browser submits the login form
- **THEN** the web service SHALL set auth cookies and return an HTML page or fragment representing the session result

#### Scenario: Logout form is submitted from the browser

- **WHEN** a browser submits the logout form
- **THEN** the web service SHALL delegate token revocation to the users service and clear browser auth cookies

### Requirement: Web exposes the documented public routes

The web service MUST expose the documented browser routes for health, auth, catalog browsing, cart interaction, and order browsing using server-rendered HTML responses where browser-facing.

#### Scenario: Public route is requested

- **WHEN** a client calls a documented public route
- **THEN** the route SHALL be served as HTML or an HTML fragment without requiring internal service access

#### Scenario: Non-browser route is requested

- **WHEN** a client calls a documented health or simulator webhook route
- **THEN** the route MAY preserve its non-HTML contract
