## Context

The repo already has the core backend primitives for unified catalog and stock creation, but they are not connected into a seller-facing browser flow. The `products` service now owns catalog rows, stock rows, and reservation behavior behind one gRPC boundary; `web` currently renders browse pages and checkout flows over `chi`, but does not expose any seller product management routes.

Recent web changes matter for this design. Protected browser routes are now grouped under middleware that extracts the authenticated user ID into request context and returns an unauthorized popup if auth is missing. Page-oriented reads such as products, cart, and orders also distinguish unavailable dependencies by rendering full unavailable pages, while interactive mutations commonly return popup or fragment responses. The seller-management flow should follow those established browser contracts instead of reintroducing redirect-based auth or a parallel UI pattern. The seller create page itself is a static form shell and does not require any downstream read before the page renders.

## Goals / Non-Goals

**Goals:**

- Add an authenticated browser flow for sellers to create a catalog product and provide its initial available quantity.
- Use the merged `products` boundary for both catalog creation and stock initialization, with the `web` service composing that flow over gRPC.
- Make product detail pages render stock-aware availability instead of a hardcoded stock placeholder.
- Adopt the v1 identity rule that `merchant_id` uses the authenticated `user_id` value for seller-owned product creation.
- Preserve the current architecture where the web edge handles forms and delegates domain work to internal services.
- Preserve the current web conventions for protected middleware, unavailable page rendering, and popup/fragment mutation responses.

**Non-Goals:**

- Introducing a separate merchant domain model, merchant memberships, or role hierarchy.
- Adding stock adjustment history, batch imports, post-create stock editing workflows, or search-index architecture changes.
- Re-splitting stock ownership into a separate service or exposing browser-facing JSON APIs.
- Redesigning reservation, checkout, or payment flows beyond the stock-aware reads needed for product display.

## Decisions

### Decision: Implement a single seller flow that creates product and initial stock in one products call

The first seller-management slice will be one browser action that submits product fields and an initial quantity. The web handler will call the merged products create method once so the catalog record and stock record are initialized inside the unified boundary.

Rationale:

- This is the smallest useful vertical slice for a marketplace seller.
- It matches the merged service boundary instead of rebuilding the old cross-service orchestration.
- It avoids a half-finished state where products exist but cannot be stocked from the browser.

Alternatives considered:

- Create product only and defer stock initialization to a second workflow. Rejected because it leaves newly created products in an unusable state for sellers.
- Add a standalone stock management tool first. Rejected because stock without product creation does not unlock marketplace onboarding.

### Decision: Use `products` as the only downstream dependency for seller listing creation

The web flow will depend on the merged `products` service for both creation and stock-aware reads.

Rationale:

- It matches the current service boundary and keeps the web integration path straightforward.
- It keeps the web dependency graph smaller and simpler.
- It keeps the current product database as the system of record and avoids introducing a search store as a primary write path.

Alternatives considered:

- Split listing creation and stock reads across separate downstream dependencies again. Rejected because it would reintroduce stale architecture into a merged system.
- Move product storage to Elasticsearch or another search-oriented store now. Rejected because the current need is canonical catalog creation, not advanced search infrastructure.

### Decision: Keep `merchant_id` as a distinct field, but source it from the authenticated user ID in v1

The web service will derive the seller's `merchant_id` from the authenticated user context and pass that value into product creation unchanged.

Rationale:

- This preserves the existing domain language in products, orders, and payment while avoiding a parallel merchant identity system.
- It leaves room for a later migration where a merchant can have multiple users or storefronts.

Alternatives considered:

- Rename domain fields from `merchant_id` to `user_id`. Rejected because it would leak the temporary v1 identity shortcut into longer-lived contracts.
- Add a new merchant service now. Rejected because it expands scope before seller listing flows exist.

### Decision: Read stock-aware product detail data from the merged products API

The web service will use products-owned stock-aware detail reads when rendering product pages.

Rationale:

- It matches the merged products contract.
- It avoids unnecessary downstream fan-out from the browser edge.

Alternatives considered:

- Reintroduce a separate stock lookup path from web. Rejected because that read now belongs inside the products boundary.
- Skip stock on product pages for now. Rejected because the current pages already model stock and seller creation should result in visible availability.

### Decision: Require explicit initial stock in the browser flow

The first seller flow will require an explicit initial stock field and will rely on the merged products contract for final validation.

Rationale:

- It matches the current products spec, which requires explicit initial stock at listing creation.
- It keeps browser validation aligned with the service contract instead of recreating stale assumptions.

Alternatives considered:

- Omit stock from the browser flow and rely on defaults. Rejected because the merged create contract requires explicit initial stock.

### Decision: Show stock-aware detail data on product detail pages only

The first implementation will fetch and display stock on product detail pages. Product list pages will remain catalog-focused and will not require exact per-item stock in this change.

Rationale:

- It is the smallest useful read path for stock-aware availability.
- It avoids adding heavier list reads or requiring batch stock APIs in the first slice.

Alternatives considered:

- Render exact stock counts on product lists. Rejected because it increases read complexity immediately.
- Render a list-page availability label. Rejected because it still requires broader stock composition for limited user value in the first slice.

### Decision: Introduce authenticated seller routes in the existing server-rendered web model

Seller management will be exposed as protected `chi` routes that render full HTML pages and handle form submissions at the browser edge, consistent with the repo's existing server-rendered and Datastar-friendly approach. The routes should plug into the current protected middleware so handlers receive user identity from request context rather than re-parsing auth state themselves.

Rationale:

- It matches the current browser architecture and avoids creating a second UI delivery style.
- It keeps auth, validation feedback, and gRPC orchestration in one place while matching the current popup-based unauthorized behavior.

Alternatives considered:

- Add an API-first management surface. Rejected because the current web capability is explicitly HTML-first.

### Decision: Reuse the current web error-surface split for seller management

Seller management pages will follow the same browser response split already used elsewhere in `services/web`: protected middleware rejects unauthenticated access before handler execution, page reads can render full unavailable pages when dependencies are down, and form submissions use popup-style error responses unless a fragment update is more appropriate.

Rationale:

- It aligns the new seller flow with current products, cart, and orders behavior.
- It avoids mixed auth behavior where some protected routes redirect and others popup.

Alternatives considered:

- Reintroduce login redirects for seller routes only. Rejected because it conflicts with the current middleware behavior and test expectations.
- Create bespoke seller-only error handling. Rejected because it would diverge from the rest of the web surface without a clear user need.

### Decision: Keep mixed-merchant carts but scope checkout to one merchant group at a time

The cart page will group items by `merchant_id` and render a separate checkout action for each merchant group. Checkout will continue to create exactly one order per submit, then remove only that merchant group's items from the cart.

Rationale:

- It preserves the existing single-order checkout flow and downstream contracts.
- It avoids introducing a multi-order transactional checkout model in the same change.
- It makes merchant boundaries visible to the buyer before checkout.

Alternatives considered:

- Split one checkout submit into multiple order creations. Rejected because it introduces partial success and cart reconciliation complexity.
- Reject mixed-merchant carts at checkout. Rejected because grouping in the cart gives a clearer buyer flow with less hidden behavior.

## Risks / Trade-offs

- [The browser flow now depends on a richer products API than the old catalog-only flow] -> Mitigate by keeping the first slice limited to listing creation and detail reads, with targeted tests for stock-aware responses.
- [Using `merchant_id == user_id` can constrain future multi-user merchants] -> Mitigate by keeping `merchant_id` as a separate field in contracts and documenting the v1 mapping as an implementation assumption, not a permanent domain model.
- [Stock-aware rendering can increase web composition cost] -> Mitigate by limiting the first slice to product detail pages, then evaluating broader read patterns only if needed.
- [Protected seller flows need authenticated user identity, not just auth presence] -> Mitigate by reusing the existing protected middleware context wiring and adding tests that cover identity propagation into gRPC create requests.
- [New seller page and cart behavior could drift from current unavailable/popup UX conventions] -> Mitigate by implementing seller routes and merchant-group checkout inside the existing handler and response helpers rather than creating route-specific response logic.

## Migration Plan

1. Add spec deltas for products, web, and server-rendered-web.
2. Expand the web products client and dependency interfaces for unified listing creation and stock-aware product reads.
3. Add protected seller pages and form handlers for product creation with explicit initial stock using the current middleware and response helpers.
4. Update product detail rendering to fetch and display products-owned stock data while preserving the existing unavailable-page behavior for page reads.
5. Group cart rendering by merchant and keep checkout merchant-scoped.
6. Add or update tests across web and service layers for create flows, protected routes, stock rendering, and seller-create unavailability.

Rollback strategy:

- Revert the web route and client wiring if seller flows prove unstable.
- Because this change primarily composes existing products APIs, rollback is mostly at the web edge; any created listing data remains valid under current domain models.

## Open Questions

- None for this revision.
