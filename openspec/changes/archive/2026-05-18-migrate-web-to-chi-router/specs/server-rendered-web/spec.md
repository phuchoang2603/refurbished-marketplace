## MODIFIED Requirements

### Requirement: Web renders server-side pages

The web service MUST render marketplace pages on the server as usable browser UI pages instead of requiring the browser to assemble them from JSON data, including auth-related pages and post-auth transitions.

#### Scenario: A page is requested directly

- **WHEN** a browser requests a top-level marketplace page
- **THEN** the web service SHALL return a complete HTML page for that route

#### Scenario: A page uses shared UI assets

- **WHEN** a browser requests a server-rendered page
- **THEN** the page SHALL include the shared web UI styling and shell needed for a cohesive marketplace experience

#### Scenario: Auth flow completes

- **WHEN** a login, registration, or logout interaction succeeds
- **THEN** the web service SHALL render or redirect into a browser flow that is immediately useful for marketplace navigation

#### Scenario: Auth interruption resumes cart flow safely

- **WHEN** a guest browses products, adds items to the cart, and is interrupted by authentication at checkout
- **THEN** the post-login browser flow SHALL return the user to a usable cart or intended page state without exposing a token-debug view or silently replaying the original checkout mutation

### Requirement: Web supports Datastar fragment updates

The web service MUST return HTML fragments or Datastar SSE patch responses that Datastar can morph into existing DOM targets for interactive browser updates.

#### Scenario: A partial interaction is submitted

- **WHEN** a browser submits a Datastar-enabled interaction
- **THEN** the web service SHALL return HTML suitable for patching the targeted DOM element

#### Scenario: A fragment response is rendered

- **WHEN** the web service returns a fragment response
- **THEN** the response SHALL include markup with stable DOM IDs that match the target used by the interaction

#### Scenario: Router migration preserves fragment behavior

- **WHEN** the router implementation changes underneath the browser edge
- **THEN** Datastar-enabled routes SHALL preserve their HTML-first interaction model without introducing browser-facing JSON APIs
