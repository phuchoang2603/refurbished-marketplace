## Why

The web service currently ships a small local UI primitive layer in `internal/views/shared` that the project owner does not want to continue designing or maintaining by hand. Adopting vendored DatastarUI components and utilities gives the server-rendered web app a more complete, reusable UI foundation while staying aligned with the existing `templ` + Tailwind + Datastar stack.

## What Changes

- Replace the current hand-rolled shared UI primitives in `services/web/internal/views/shared/ui.*` with a new vendored DatastarUI-based component foundation.
- Migrate existing server-rendered web pages to direct component usage style instead of preserving the current shared wrapper API.
- Preserve the existing HTML-first, Datastar-enabled interaction model and marketplace shell while reworking the component and styling foundation underneath it.
- Vendor selected DatastarUI component packages and supporting utility packages into the repo instead of cloning the upstream demo application wholesale.
- Remove the legacy local UI primitive layer after all existing pages have been migrated to the new foundation.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `basic-webui-frontend`: the shared browser UI foundation changes from local custom primitives to a vendored DatastarUI-based direct component model.
- `server-rendered-web`: server-rendered pages continue to render HTML-first responses, but their shared styling and interactive component foundation changes to vendored DatastarUI assets and components.

## Impact

- Affected code: `services/web/internal/views/**`, shared shell/layout assets, Tailwind configuration, and any vendored UI/component utility packages added under `services/web`.
- Dependencies: web UI code will depend on vendored DatastarUI component and utility packages, plus any supporting Tailwind/theme conventions they require.
- Systems: this is isolated to the browser edge and server-rendered presentation layer; product, cart, auth, and internal gRPC service boundaries remain unchanged.
