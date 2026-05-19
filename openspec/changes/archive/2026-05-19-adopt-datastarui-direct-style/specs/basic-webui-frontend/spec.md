## MODIFIED Requirements

### Requirement: Web UI provides a cohesive marketplace shell

The web service MUST provide a cohesive browser UI shell for marketplace pages, including shared navigation, typography, spacing, and responsive layout suitable for desktop and mobile browsers, and it MUST build repeated UI primitives from the vendored DatastarUI-based component foundation using direct component usage style.

#### Scenario: A marketplace page is rendered

- **WHEN** a browser requests a marketplace page
- **THEN** the page SHALL render within the shared marketplace shell with navigation to catalog, cart, orders, and auth flows

#### Scenario: A page is viewed on a narrow screen

- **WHEN** a browser renders the web UI on a mobile-width viewport
- **THEN** the layout SHALL remain usable without horizontal scrolling for primary content and controls

#### Scenario: Shared primitives are rendered

- **WHEN** a page renders repeated UI primitives such as buttons, fields, cards, selects, dialogs, or empty states
- **THEN** the page SHALL use the vendored DatastarUI-based server-rendered components directly instead of the legacy local shared primitive layer
