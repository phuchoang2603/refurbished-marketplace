## Context

The web service is the public browser edge for the marketplace, but many browser pages currently depend on successful gRPC calls to downstream domain services during initial render or interactive updates. This creates a brittle user experience where a single service outage can make the whole UI feel unavailable even though the web shell, navigation, auth pages, and unrelated features could still function.

The current code shape already separates public browser routes, protected browser routes, and non-browser routes in `services/web/internal/handlers`. The missing piece is a consistent dependency policy: which routes must hard-fail when a backing service is unavailable, which routes may render partial or fallback UI, and where localized browser feedback should appear.

## Goals / Non-Goals

**Goals:**

- Keep the web shell, navigation, auth pages, and non-domain routes available when one downstream service is down.
- Localize downstream failures to the feature being used instead of turning the whole web edge into a failure surface.
- Preserve gRPC composition at the web edge while making page handlers explicit about soft-fail versus hard-fail behavior.
- Keep interactive browser responses HTML-first by using existing popup, fragment, and full-page rendering helpers.

**Non-Goals:**

- Replacing gRPC with another composition mechanism.
- Redesigning internal service APIs or ownership boundaries.
- Building a full client-side offline mode or caching layer.
- Changing non-browser webhook or health-check contracts beyond keeping them outside browser feature failures.

## Decisions

### Decision: classify browser routes by dependency criticality

The web edge will classify routes and handlers into three groups:

- offline-safe routes that must remain available without downstream domain data, such as shell, auth pages, and health routes;
- soft-fail feature routes that may render fallback UI or show localized browser feedback when a domain service is unavailable;
- hard-fail mutations that must report an error because the requested action cannot complete safely.

Rationale: this keeps the public edge stable without pretending that write operations or service-owned data can succeed while their owners are down.

Alternatives considered:

- Hard-fail every route when any dependency is unhealthy: simplest operationally, but it defeats the purpose of a resilient browser edge.
- Fully cache all downstream data at the web edge: improves resilience but introduces ownership drift, staleness, and larger scope than needed.

### Decision: degrade at the handler and view boundary, not in shared transport plumbing

The web service will keep its existing gRPC clients, but each browser handler will decide whether a given downstream failure should produce fallback HTML, a popup, or a hard error response. Shared helpers should support those outcomes, but the choice belongs to the feature handler.

Rationale: different routes have different acceptable fallback behavior. Product browsing can show an unavailable state, while checkout must fail explicitly.

Alternatives considered:

- Global middleware that swallows all gRPC failures: too coarse and would hide important domain-specific distinctions.
- Service discovery or connection-level readiness gates that block startup until all dependencies are healthy: preserves the current coupling instead of reducing it.

### Decision: keep the shell and navigation renderable without feature data

The shared marketplace shell and route registration should remain renderable even if products, cart, or orders data cannot be fetched. Feature pages may show empty states, unavailable sections, or popups, but the browser should still receive a valid page or fragment response where feasible.

Rationale: users should be able to navigate, sign in, sign out, and discover which feature is impaired without losing the entire web experience.

Alternatives considered:

- Server-side redirects to a global outage page: reduces clarity about which feature failed and makes unrelated flows unavailable.

### Decision: use inline fallback for read paths and popups for action failures

Browser page loads and other read-path failures will use inline unavailable states inside the rendered page. User-triggered actions and mutations will use localized popup or inline action feedback when the backing service is unavailable.

Rationale: this keeps page rendering calm and predictable while still surfacing action failures immediately without adding extra page-specific branching.

Alternatives considered:

- Popups for all failures: simpler at first, but noisy and awkward for normal page loads.
- Distinct handling per feature page: more flexible, but adds unnecessary UI branching for the initial implementation.

### Decision: allow the web process to boot without healthy downstream services

The web process will tolerate missing downstream client targets at startup as long as its own configuration is valid. Downstream dependency failures will be handled per request using the same degraded-mode behavior used for runtime outages.

Rationale: one availability model for both startup and runtime is simpler than maintaining separate boot-time and request-time rules.

Alternatives considered:

- Require all downstream services to be reachable before startup: preserves the current coupling and makes the whole browser edge unavailable during partial outages.

### Decision: use one shared unavailable-state component first

Product, cart, and orders pages will use one shared unavailable-state component with small copy variations such as title and message, rather than distinct layouts per feature.

Rationale: this is the smallest implementation that keeps the browser experience consistent and easy to maintain.

Alternatives considered:

- Distinct unavailable layouts per feature: richer, but larger scope and harder to keep consistent.

### Decision: keep the first unavailable-state implementation copy-only

The shared unavailable-state component will use copy-only feedback in the first implementation and will not include retry affordances or other extra controls.

Rationale: copy-only feedback is the smallest path to graceful degradation and avoids extra interaction design while the fallback model is still being established.

Alternatives considered:

- Add retry controls immediately: useful in some cases, but increases UI and handler scope beyond the minimal first implementation.

### Decision: keep hard-fail semantics for service-backed mutations

Mutations such as login, registration, logout, checkout, and order creation will continue to fail when their owning service is unavailable. The improvement is that the failure should be surfaced as localized browser feedback rather than broad router or shell failure.

Rationale: these operations require the downstream service to succeed and cannot be meaningfully emulated at the web edge.

Alternatives considered:

- Queueing browser mutations for replay: adds complexity and changes user expectations around immediate consistency.

## Risks / Trade-offs

- [Inconsistent fallback behavior across handlers] -> Define route categories and test expectations per feature so handlers degrade in predictable ways.
- [Overuse of popups can feel noisy] -> Prefer inline unavailable states for page renders, and reserve popups for interactive failures or protected actions.
- [The web edge may still block at startup if client construction is strict] -> Separate startup readiness from per-request downstream availability and allow the web process to serve routes even when a client target is temporarily unhealthy.
- [Partial pages may hide real incidents] -> Keep health routes and operational observability intact so degraded UI behavior does not obscure service outages.

## Migration Plan

1. Update specs to classify web-edge behavior around downstream outages.
2. Refactor web handlers so page renders and fragment handlers explicitly choose fallback or hard-fail behavior.
3. Adjust tests to verify localized failures instead of whole-edge redirects or broad unavailability.
4. Roll out incrementally by feature area, starting with public pages and cart/orders routes.
5. Roll back by restoring direct hard-fail behavior in affected handlers if fallback rendering introduces incorrect UI states.

## Open Questions

- None for the initial implementation.
