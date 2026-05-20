## MODIFIED Requirements

### Requirement: Web keeps internal composition over gRPC

The web service MUST continue to compose data from internal services over gRPC while rendering browser responses, and it MUST localize downstream read-path failures to the affected page, fragment, or interaction when a usable browser response can still be produced.

#### Scenario: A rendered page needs marketplace data

- **WHEN** a page requires data from internal domain services
- **THEN** the web service SHALL fetch that data through the existing gRPC clients before rendering HTML

#### Scenario: A page dependency is unavailable

- **WHEN** a browser page depends on a downstream service that is unavailable
- **THEN** the web service SHALL return a usable HTML page with inline localized unavailable-state content, partial content, or other feature-scoped fallback instead of failing the entire browser shell when that route can degrade safely

#### Scenario: A fragment dependency is unavailable

- **WHEN** a Datastar fragment or interactive browser update depends on a downstream service that is unavailable
- **THEN** the web service SHALL return localized browser feedback or a fallback fragment scoped to that feature instead of turning unrelated browser features unavailable
