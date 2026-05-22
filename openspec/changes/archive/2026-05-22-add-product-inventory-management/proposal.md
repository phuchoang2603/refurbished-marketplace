## Why

The marketplace can browse catalog entries and reserve stock during checkout, but it still lacks a seller-facing browser flow for creating listings. This change adds that missing management path on top of the merged catalog boundary so authenticated users can create products with initial stock through `services/web` without bypassing the web service.

## What Changes

- Add browser routes and server-rendered management pages for authenticated users to create products with explicit initial stock through the merged `products` service within the current protected route structure.
- Extend web-to-service composition so the web service can call unified product creation and stock-aware product read gRPC methods.
- Define the v1 identity rule that the authenticated seller's `user_id` and `merchant_id` use the same value when creating and owning catalog entries.
- Render product availability from products-owned stock data on product detail pages instead of placeholder stock values.
- Group cart items by `merchant_id` on the cart page and keep checkout scoped to one merchant group per submit so the existing single-order checkout flow can be reused.
- Keep browser-side quantity validation aligned with the merged products contract by requiring explicit initial stock in the seller flow.
- Reuse the current web response conventions for seller flows: protected routes rely on middleware-provided user identity, page reads render HTML/unavailable states, and mutations return browser-friendly popup or fragment responses.
- Non-goal: introducing a separate merchant account model, multi-user merchant teams, standalone stock adjustment workflows beyond initial stock creation, or search-index storage changes such as moving the catalog source of truth to Elasticsearch.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `products`: unified catalog requirements expand to cover authenticated seller-owned listing creation and stock-aware detail reads used by the web flow.
- `web`: browser-edge requirements expand to cover authenticated seller product management forms and the v1 `merchant_id == user_id` identity mapping.
- `server-rendered-web`: rendered page requirements expand to cover seller-facing product creation UI and stock-aware product pages.
- `cart`: rendered cart requirements expand to cover merchant-grouped cart presentation and merchant-scoped checkout actions.

## Impact

- Affected code: `services/web` handlers, middleware, views, client registry, and route wiring; `services/products` gRPC client usage and tests.
- APIs: web will consume unified products create/read methods; product detail page composition will use products-owned stock-aware reads.
- Dependencies: web extends its products service dependency for unified create and stock-aware read behavior.
- Systems: seller onboarding and catalog management begin at the browser edge, while preserving the merged boundary where `products` owns catalog, stock, and reservation state.
