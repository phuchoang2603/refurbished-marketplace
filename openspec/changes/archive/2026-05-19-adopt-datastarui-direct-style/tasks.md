## 1. Vendor DatastarUI Foundation

- [x] 1.1 Identify the current shared primitives and Datastar behaviors used by `services/web/internal/views/**` and map them to the DatastarUI components and utility packages that need to be vendored
- [x] 1.2 Vendor the selected DatastarUI component packages into a `services/web` package layout that stays as close to upstream DatastarUI structure and naming as practical
- [x] 1.3 Vendor the required DatastarUI utility packages and supporting helpers used by the selected components

## 2. Align Shared Styling and Theme Setup

- [x] 2.1 Update `services/web` Tailwind configuration to include the vendored component source paths and any required theme or plugin conventions
- [x] 2.2 Replace the shared CSS foundation so vendored DatastarUI components render with the expected tokens, base styles, and interactive states without preserving the legacy marketplace styling
- [x] 2.3 Update the shared shell template to load the DatastarUI-aligned styling foundation while preserving only marketplace-specific structure, navigation, and Datastar runtime wiring

## 3. Migrate Page Templates to Direct Component Usage

- [x] 3.1 Migrate auth page templates from the legacy shared primitive layer to direct imports of the vendored component packages
- [x] 3.2 Migrate product and cart page templates from the legacy shared primitive layer to direct imports of the vendored component packages
- [x] 3.3 Migrate orders and any remaining server-rendered page templates to the vendored direct component style

## 4. Remove Legacy UI Layer and Verify Behavior

- [x] 4.1 Delete the obsolete `services/web/internal/views/shared/ui.*` primitive layer and remove any unused helper code left behind by the migration
- [x] 4.2 Update or add tests that cover full-page rendering and Datastar fragment responses affected by the component migration
- [x] 4.3 Run the relevant `services/web` test suite and verify the main browser flows still render and behave correctly with the vendored DatastarUI foundation
