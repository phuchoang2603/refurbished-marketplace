## Context

The current web edge uses `http.NewServeMux()` in `services/web/cmd/web/main.go` and registers all browser and non-browser routes from a single `Register(mux *http.ServeMux)` method in `services/web/internal/handlers/handler.go`. Auth concerns are layered manually with `requireAccessToken(...)` and `withViewAuth(...)`, which works today but is already flattening multiple route concerns into one place.

The browser auth flow is also skewed toward debugging rather than normal marketplace navigation. `handleLogin` currently sets cookies and renders `views.TokensPage(...)`, and logout renders message pages instead of treating auth as a short transition back into the marketplace shell. The project already expects server-rendered HTML and Datastar-compatible fragments, so the router migration should serve that interaction model rather than introduce a new one.

## Goals / Non-Goals

**Goals:**

- Introduce Chi as the web router so route groups and middleware composition are explicit.
- Separate public, authenticated, and non-browser/webhook routes into clearer route trees.
- Make login, register, and logout usable browser flows that return users to meaningful pages or fragments.
- Introduce request-level OpenTelemetry middleware at the web edge so browser requests enter `services/web` with tracing context.
- Preserve the existing gRPC composition model and HTML/Datastar browser contract.

**Non-Goals:**

- Add admin inventory workflow or any new inventory client to `services/web`.
- Add tracing/debug routes in this change.
- Add gRPC tracing interceptors or Kafka/outbox trace propagation in this change.
- Replace `templ` or Datastar with a different rendering model.
- Redesign the whole visual system beyond what is needed to make auth and navigation usable.

## Decisions

### Decision: Migrate only the web service router to Chi

Chi will be introduced in `services/web` only. Other services will continue using their current transport setup unless a future change justifies broader router standardization.

Rationale:

- The routing pain and middleware grouping problem exist at the browser edge, not across the gRPC services.
- This keeps the change scoped to the place where route organization and browser UX actually matter.

Alternatives considered:

- Keep `http.ServeMux` and manually refactor the register function. Rejected because it preserves the flat routing structure that is already becoming awkward.
- Migrate all services to Chi. Rejected because the current need is route composition in `web`, not cross-service router uniformity.

### Decision: Organize routes by browser concern, not by handler file

The Chi tree should group routes into public browser routes, authenticated browser routes, and non-browser endpoints such as health and simulator webhook handling.

Rationale:

- Auth and browser shell behavior are applied by route concern, not by file location.
- This makes middleware placement and future route additions clearer.

Alternatives considered:

- Keep one large route registration function under Chi. Rejected because it would move libraries without improving structure.

### Decision: Add tracing as Chi middleware at the web edge only

OpenTelemetry tracing should be introduced through Chi middleware in `services/web` so incoming browser requests create or continue request-scoped trace context inside the web edge. This change will stop at the web router boundary and will not include downstream gRPC interceptors or Kafka propagation.

Rationale:

- Chi gives the web edge a natural place to apply request middleware once instead of hand-wrapping handlers.
- Web-edge tracing is useful immediately for browser request visibility even before deeper distributed propagation is added.
- Keeping tracing scoped to the web edge avoids expanding this router migration into a full cross-service tracing program.

Alternatives considered:

- Skip tracing entirely in this migration. Rejected because Chi middleware is a natural opportunity to add web request instrumentation with limited extra scope.
- Add gRPC interceptors and async trace propagation now. Rejected because that would broaden the change well beyond the web router migration.

### Decision: Replace token/session landing pages with browser-oriented redirects

Successful login and registration should return users to a useful browser destination such as the intended protected `GET` route, the catalog, or the cart, rather than showing a token/session detail page. Logout should clear cookies and return the user to a browser-appropriate page or fragment.

Rationale:

- The current flow exposes implementation-shaped session details instead of a usable marketplace journey.
- Browser auth should feel like navigation, not like an API demo.

Alternatives considered:

- Keep the token/session page and only restyle it. Rejected because the main issue is flow, not cosmetics.

### Decision: Preserve `next` for protected `GET` requests but do not replay protected `POST` actions

When an unauthenticated user is interrupted on a protected `GET` request, the web edge should preserve the intended destination and return the user there after successful login. When an unauthenticated user hits a protected `POST` such as checkout, the web edge should redirect them to login with a safe resume destination such as `/cart` rather than attempting to replay the original mutation automatically.

Rationale:

- Guests can already browse products and build a cart before authentication, so `/cart` is a natural resume point after auth interruption.
- Preserving `GET` destinations improves UX without introducing mutation replay complexity.
- Replaying protected `POST` actions would require carrying form state and mutation intent across auth, which is more complexity than this router migration needs.

Alternatives considered:

- Always redirect to `/products` after login. Rejected because it drops useful context when a user was trying to reach a specific protected page.
- Replay the original protected `POST` after login. Rejected for this stage because it adds state capture and idempotency complexity.

### Decision: Preserve Datastar-compatible forms and fragment responses

The migration should preserve the HTML-first browser contract and Datastar-compatible fragments. Chi should be used to simplify route grouping, not to switch the response model.

Rationale:

- The repository already documents server-rendered pages and fragment updates as the browser model.
- Router migration should reduce structural friction without creating a rendering migration.

Alternatives considered:

- Introduce JSON endpoints for auth/cart interactions during the migration. Rejected because it conflicts with the current web edge contract.

## Risks / Trade-offs

- [Router migration changes subtle route behavior] -> Keep route coverage explicit in web tests, especially auth, cart, orders, and webhook paths.
- [Auth UX cleanup expands into a full visual redesign] -> Limit UI changes to the auth and navigation experience needed to make the browser flow usable.
- [Datastar fragment behavior regresses during route reshaping] -> Preserve stable route paths and fragment targets while moving only the routing/middleware layer.
- [Chi adds dependency and pattern overhead without enough payoff] -> Keep the migration confined to `services/web` and use route groups/middleware consistently enough to justify the dependency.
- [Tracing scope expands into full distributed propagation work] -> Limit this change to Chi middleware in `services/web` and defer gRPC/Kafka propagation to a follow-up change.

## Migration Plan

1. Add Chi to `services/web` and replace the top-level `ServeMux` setup with a Chi router.
2. Reorganize route registration into public, authenticated, and non-browser groups while preserving current route paths.
3. Introduce request-scoped OpenTelemetry middleware in the Chi stack for the web edge.
4. Update auth handlers and views so successful login/register/logout produce usable browser navigation outcomes.
5. Re-run and extend web tests to cover route behavior, protected `GET` destination preservation, and safe protected `POST` resume behavior.

Rollback strategy:

- Revert the `services/web` router wiring and auth-flow view changes together, since the main user-visible change is at the web edge.

## Open Questions

None at this stage.
