## 1. Web Dependency Wiring

- [x] 1.1 Expand the products gRPC client usage in `services/web/internal/clients` and register any needed unified create/read wiring in web startup.
- [x] 1.2 Expand the web handler dependency interfaces to support unified product creation and stock-aware product reads.
- [x] 1.3 Update web test fakes and route test setup to satisfy the new dependency surface.

## 2. Seller Product Creation Flow

- [x] 2.1 Add protected `chi` routes for seller product management pages and form submissions in `services/web/internal/handlers`, reusing the existing protected middleware group.
- [x] 2.2 Implement the create-product form handler so it derives `merchant_id` from the authenticated user ID in request context and calls the unified create method in the products service with the initial stock quantity.
- [x] 2.3 Keep seller-form stock validation aligned with the merged products contract by requiring explicit initial stock input.
- [x] 2.4 Add server-rendered views for the seller product creation page and success/error states using the existing shared web shell and current popup/fragment response helpers.

## 3. Stock-Aware Product Rendering

- [x] 3.1 Update product detail handlers to fetch stock-aware product data from the products service when rendering product detail pages.
- [x] 3.2 Keep product list rendering catalog-only in this change instead of adding exact per-item stock reads.
- [x] 3.3 Handle missing stock data or downstream lookup failures on product detail pages with the existing unavailable-page behavior for page reads and browser-friendly mutation responses where applicable.

## 4. Verification

- [x] 4.1 Add or update flow-oriented web tests covering protected seller routes, `merchant_id == user_id` request mapping, the unified create flow, and seller-create unavailability.
- [x] 4.2 Add or update flow-oriented tests covering stock-aware product detail rendering behavior, catalog-focused list behavior, and merchant-scoped cart checkout flow.
- [x] 4.3 Run the relevant service test suites for `services/web` and `services/products` and address any failures needed for this change.
