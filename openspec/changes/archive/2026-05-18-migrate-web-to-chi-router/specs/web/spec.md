## MODIFIED Requirements

### Requirement: Web owns the public browser edge

The web service MUST own the public browser surface, authorization boundary, browser auth cookies, and browser-facing UI routes, and it MUST organize those routes so public, authenticated, and non-browser concerns can apply middleware consistently.

#### Scenario: A protected route is called

- **WHEN** a client calls a protected browser endpoint
- **THEN** the web service SHALL validate the browser auth cookie or access token and forward trusted identity to internal services

#### Scenario: A browser form is submitted

- **WHEN** a browser submits a web UI form
- **THEN** the web service SHALL process the form at the browser edge and translate successful actions into internal gRPC calls

#### Scenario: A non-browser route is called

- **WHEN** a client calls a health or simulator webhook route
- **THEN** the web service SHALL keep that route outside browser-auth middleware and preserve its documented non-browser contract

#### Scenario: A browser request enters the router

- **WHEN** a browser request enters the web router
- **THEN** the web service SHALL apply request-scoped OpenTelemetry middleware at the web edge so handlers execute with tracing context available on the request

### Requirement: Web delegates auth session logic to users

The web service MUST delegate login and logout session logic to the users service while presenting browser-friendly auth UI responses and managing browser cookie persistence.

#### Scenario: Login request arrives

- **WHEN** a client calls the login endpoint
- **THEN** the web service SHALL invoke the users service for session issuance

#### Scenario: Login form is submitted from the browser

- **WHEN** a browser submits the login form
- **THEN** the web service SHALL set auth cookies and return an HTML page, fragment, or redirect that moves the user into a usable marketplace flow instead of a token-debug landing page

#### Scenario: Protected page redirects through login

- **WHEN** an unauthenticated browser is redirected away from a protected `GET` route
- **THEN** the web service SHALL preserve the intended destination and return the user to that page after successful login when it is safe to do so

#### Scenario: Protected mutation redirects through login

- **WHEN** an unauthenticated browser is redirected away from a protected `POST` route such as checkout
- **THEN** the web service SHALL return the user to a safe resume page such as `/cart` after successful login instead of replaying the mutation automatically

#### Scenario: Logout form is submitted from the browser

- **WHEN** a browser submits the logout form
- **THEN** the web service SHALL delegate token revocation to the users service, clear browser auth cookies, and return an HTML page, fragment, or redirect that leaves the browser in a usable signed-out state
