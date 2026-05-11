## MODIFIED Requirements

### Requirement: Web renders server-side pages

The web service MUST render marketplace pages on the server as usable browser UI pages instead of requiring the browser to assemble them from JSON data.

#### Scenario: A page is requested directly

- **WHEN** a browser requests a top-level marketplace page
- **THEN** the web service SHALL return a complete HTML page for that route

#### Scenario: A page uses shared UI assets

- **WHEN** a browser requests a server-rendered page
- **THEN** the page SHALL include the shared web UI styling and shell needed for a cohesive marketplace experience

### Requirement: Web supports Datastar fragment updates

The web service MUST return HTML fragments or Datastar SSE patch responses that Datastar can morph into existing DOM targets for interactive browser updates.

#### Scenario: A partial interaction is submitted

- **WHEN** a browser submits a Datastar-enabled interaction
- **THEN** the web service SHALL return HTML suitable for patching the targeted DOM element

#### Scenario: A fragment response is rendered

- **WHEN** the web service returns a fragment response
- **THEN** the response SHALL include markup with stable DOM IDs that match the target used by the interaction

#### Scenario: An interaction updates multiple targets

- **WHEN** a Datastar-enabled interaction needs to update multiple DOM targets, signals, redirect state, or progressive UI state
- **THEN** the web service SHALL use Datastar-compatible SSE patch responses rather than introducing browser-facing JSON APIs

### Requirement: Web keeps internal composition over gRPC

The web service MUST continue to compose data from internal services over gRPC while rendering browser responses.

#### Scenario: A rendered page needs marketplace data

- **WHEN** a page requires data from internal domain services
- **THEN** the web service SHALL fetch that data through the existing gRPC clients before rendering HTML
