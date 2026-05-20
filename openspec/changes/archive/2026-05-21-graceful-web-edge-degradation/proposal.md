## Why

The web service currently depends on multiple downstream gRPC services to render key browser pages, which makes the whole UI feel unavailable when a single service is down. We need the web edge to stay usable as the public entrypoint and degrade failures at the feature boundary instead of failing the entire browser experience.

## What Changes

- Update the web edge requirements so the browser shell, navigation, auth pages, and non-domain routes remain available even when individual downstream services are unavailable, including during web-service startup.
- Define graceful degradation rules for server-rendered browser pages and interactive fragments when products, cart, orders, or similar domain services fail.
- Require feature-specific UI errors, shared inline unavailable states for page reads, and localized popups for action failures instead of broad web-edge unavailability when one downstream dependency is down.
- Clarify that service-backed mutations may still fail hard, but those failures must be localized to the feature being used.
- Non-goal: redesign internal domain APIs or remove gRPC composition from the web service.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `web`: change browser-edge availability requirements so downstream service outages do not take down unrelated browser routes or the shared shell.
- `server-rendered-web`: change rendering requirements so browser pages and fragments degrade gracefully when a backing service is unavailable.
- `basic-webui-frontend`: change UI behavior requirements so feature navigation remains usable and localized browser feedback is shown when a dependent service is down.

## Impact

- Affected code: `services/web/internal/handlers`, web gRPC client usage, page/fragment rendering paths, and browser error handling helpers.
- Affected systems: web edge, products service, cart service, orders service, users service, and payment-adjacent browser flows.
- Dependencies: no new platform dependency is required, but the web edge will need clearer startup and runtime boundaries around optional vs required downstream service calls plus a shared unavailable-state UI path.
