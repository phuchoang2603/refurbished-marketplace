## 1. Router Migration

- [x] 1.1 Add Chi to `services/web` and replace the top-level `http.NewServeMux()` setup in `cmd/web` with a Chi router.
- [x] 1.2 Reorganize route registration into public browser routes, authenticated browser routes, and non-browser/webhook routes while preserving existing URL paths.
- [x] 1.3 Move auth-related middleware application to route groups so browser auth and non-browser routes are separated cleanly.
- [x] 1.4 Add request-scoped OpenTelemetry middleware to the Chi router in `services/web` without expanding this change into gRPC or Kafka tracing propagation.

## 2. Auth Flow UX

- [x] 2.1 Replace the current login success token/session landing flow with browser-friendly redirects that preserve protected `GET` destinations when available.
- [x] 2.2 Update logout behavior and related views so signed-out users land in a usable browser state instead of a debug-style message flow.
- [x] 2.3 Implement safe auth interruption handling for protected `POST` flows such as checkout by resuming at `/cart` instead of replaying the mutation.
- [x] 2.4 Review registration and auth shell/navigation markup so login, register, and logout are discoverable and usable from the browser.

## 3. Datastar And Route Behavior

- [x] 3.1 Ensure existing Datastar-enabled forms and fragment responses still use stable targets and HTML-first responses after the router migration.
- [x] 3.2 Keep health and Stripe simulator webhook routes outside browser-auth middleware and preserve their current non-browser contracts.

## 4. Verification

- [x] 4.1 Update or add `services/web/tests` coverage for route registration behavior, auth flows, protected `GET` destination preservation, and safe protected `POST` resume behavior.
- [x] 4.2 Run the web service test suite and resolve any regressions introduced by the router and auth UX changes.
