## ADDED Requirements

### Requirement: Web renders server-side pages

The web service MUST render marketplace pages on the server instead of requiring the browser to assemble them from JSON data.

#### Scenario: A page is requested directly

- **WHEN** a browser requests a top-level marketplace page
- **THEN** the web service SHALL return HTML for that page

### Requirement: Web supports Datastar fragment updates

The web service MUST return HTML fragments that Datastar can morph into the existing DOM for interactive updates.

#### Scenario: A partial interaction is submitted

- **WHEN** a browser submits a Datastar-enabled interaction
- **THEN** the web service SHALL return HTML suitable for patching the targeted DOM element

### Requirement: Web keeps internal composition over gRPC

The web service MUST continue to compose data from internal services over gRPC while rendering browser responses.

#### Scenario: A rendered page needs marketplace data

- **WHEN** a page requires data from internal domain services
- **THEN** the web service SHALL fetch that data through the existing gRPC clients before rendering HTML
