## MODIFIED Requirements

### Requirement: Web UI provides a cohesive marketplace shell

The web service MUST provide a cohesive browser UI shell for marketplace pages, including shared navigation, typography, spacing, and responsive layout suitable for desktop and mobile browsers, and it MUST build repeated UI primitives from the vendored DatastarUI-based component foundation using direct component usage style. The shared shell and feature navigation MUST remain usable even when one feature's backing service is unavailable.

#### Scenario: A marketplace page is rendered

- **WHEN** a browser requests a marketplace page
- **THEN** the page SHALL render within the shared marketplace shell with navigation to catalog, cart, orders, and auth flows

#### Scenario: A page is viewed on a narrow screen

- **WHEN** a browser renders the web UI on a mobile-width viewport
- **THEN** the layout SHALL remain usable without horizontal scrolling for primary content and controls

#### Scenario: Shared primitives are rendered

- **WHEN** a page renders repeated UI primitives such as buttons, fields, cards, selects, dialogs, or empty states
- **THEN** the page SHALL use the vendored DatastarUI-based server-rendered components directly instead of the legacy local shared primitive layer

#### Scenario: One feature is unavailable

- **WHEN** one browser feature depends on a downstream service that is unavailable
- **THEN** the shell and navigation SHALL remain usable so the user can continue to unrelated marketplace areas

## ADDED Requirements

### Requirement: Web UI localizes downstream feature failures

The web service MUST present downstream service failures as localized browser feedback at the feature boundary instead of presenting the entire marketplace UI as unavailable.

#### Scenario: A read-only feature fails

- **WHEN** a browser navigates to a page whose backing service is unavailable but the page can still render a meaningful fallback
- **THEN** the web service SHALL render a shared inline unavailable-state component, partial content, or other feature-scoped fallback within that page

#### Scenario: A feature action fails

- **WHEN** a browser triggers an action whose backing service is unavailable
- **THEN** the web service SHALL show browser feedback such as a popup or inline error scoped to that action

#### Scenario: Different feature pages degrade consistently

- **WHEN** products, cart, or orders pages each need to show a read-path unavailable state
- **THEN** the web service SHALL reuse a shared unavailable-state component with copy variations instead of requiring distinct failure layouts per feature
