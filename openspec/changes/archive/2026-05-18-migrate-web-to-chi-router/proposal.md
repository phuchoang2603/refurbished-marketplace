## Why

The current `services/web` router is a single `http.ServeMux` registration block with hand-wrapped auth helpers, which is still workable but is already making the browser edge harder to extend. At the same time, the current login/logout browser flow is awkward enough to feel unusable, because successful auth ends on a token/session page instead of returning users to a meaningful marketplace path.

## What Changes

- Replace the current `http.ServeMux` route registration in `services/web` with a Chi-based route tree that groups public, authenticated, and non-browser endpoints more cleanly.
- Restructure the browser auth flow so login, registration, and logout feel like usable marketplace navigation flows instead of debug-style token/session pages, including preserving intended protected `GET` destinations and resuming protected `POST` interruptions at a safe page such as `/cart`.
- Introduce Chi middleware hooks for OpenTelemetry tracing in `services/web` so browser requests enter the web edge with request-scoped tracing context.
- Preserve the current HTML-first and Datastar-compatible browser contract while making fragment routes and auth middleware composition easier to reason about.
- Keep scope limited to the web service router, browser UX cleanup, and web-edge tracing middleware; do not add admin inventory workflow, gRPC interceptor propagation, Kafka/outbox tracing propagation, or new non-browser APIs in this change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `web`: update browser-edge requirements so auth routes behave as usable browser flows and route organization can support grouped middleware without changing the public ownership boundary.
- `server-rendered-web`: update fragment/page requirements so the post-auth browser experience and Datastar-compatible interactions stay coherent after the router migration.

## Impact

- Affected code is concentrated in `services/web`, especially route registration, auth handlers, middleware application, and server-rendered views.
- Adds a new router dependency and associated middleware patterns to the web service only, including request-level tracing middleware.
- Browser-facing behavior for login, logout, and related redirects/fragments will change.
- Non-goals: admin inventory tooling, tracing/debug dashboards, gRPC tracing interceptors, Kafka/outbox trace propagation, or changes to internal service contracts outside what the web edge already consumes.
