## Why

The web service already owns server-rendered browser responses, but the current pages are mostly scaffolding and expose raw service data rather than a coherent marketplace UI. A basic web UI gives the browser edge a usable first experience without introducing a separate SPA or `clients/` frontend.

## What Changes

- Improve the `services/web` catalog, product detail, cart, auth, and order pages into a basic cohesive marketplace UI.
- Add shared styling/static assets for the web edge and copy DatastarUI-inspired `templ` primitives into the app for broad use.
- Improve Datastar-driven interactions for common flows such as adding to cart, editing cart quantities, and auth form feedback, using `datastar-go` for SSE patches where interactions update DOM targets or signals.
- Persist browser auth tokens in HTTP cookies so protected browser routes do not require manual `Authorization` headers.
- Compose cart rows with product details from the products service for a usable cart display while keeping final order totals authoritative in order creation.
- Keep browser requests form/HTML-first and preserve internal gRPC composition.
- Do not introduce a separate frontend application under `clients/`.

## Capabilities

### New Capabilities

- `basic-webui-frontend`: Basic server-rendered marketplace UI, including cohesive page layout, static styling, and Datastar-compatible browser interactions.

### Modified Capabilities

- `server-rendered-web`: Strengthen existing server-rendered page and fragment requirements from technical HTML rendering to a usable browser UI experience.
- `web`: Clarify public browser route behavior for form-based auth, catalog browsing, cart interaction, and protected order flows.

## Impact

- `services/web/internal/views/` templates and view models.
- `services/web/internal/handlers/` fragment/page response behavior where needed.
- `services/web/static/` or equivalent static asset serving if shared CSS/assets are added.
- Browser auth cookie handling in `services/web/internal/auth/` and auth handlers.
- Cart page composition through existing products gRPC clients.
- Copied DatastarUI-inspired `templ` components under `services/web/internal/views/` rather than an installed component-library dependency.
- Web module dependency on `datastar-go` for Datastar SSE and patch response helpers.
- Users token response contract includes refresh-token expiry so browser refresh cookies can follow users service token lifetimes.
- `services/web/tests/` coverage for rendered pages and interactive fragments.
