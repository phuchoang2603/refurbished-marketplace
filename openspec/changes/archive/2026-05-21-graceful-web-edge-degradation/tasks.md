## 1. Route and dependency policy

- [x] 1.1 Inventory `services/web` browser routes and classify each as offline-safe, soft-fail, or hard-fail.
- [x] 1.2 Update web-edge route registration and shared handler helpers so public shell routes remain available independently of protected or feature-specific failures.
- [x] 1.3 Separate web process startup/readiness from downstream feature availability so the UI can boot even when one service target is unavailable.

## 2. Public page degradation

- [x] 2.1 Add or reuse one shared inline unavailable-state component and helper flow for read-path failures in templ-rendered pages.
- [x] 2.2 Refactor product listing and product detail handlers to return usable fallback HTML when products data cannot be fetched.
- [x] 2.3 Refactor cart page rendering to tolerate missing composed product details and show the shared inline unavailable state instead of failing the whole page.

## 3. Protected feature degradation

- [x] 3.1 Refactor orders pages to localize downstream read failures with the shared inline unavailable state while preserving the shared shell and authenticated browser flow.
- [x] 3.2 Keep protected mutations such as checkout and order creation as hard-fail actions, but ensure failures surface as feature-scoped popups or inline action feedback.
- [x] 3.3 Review auth-dependent browser flows so authentication failures and downstream domain failures remain distinct in handler behavior and UI feedback.

## 4. Verification

- [x] 4.1 Update `services/web/tests` route and handler coverage to assert localized degradation behavior instead of whole-edge unavailability.
- [x] 4.2 Add tests for at least one public page failure, one protected page failure, and one hard-fail mutation when a downstream service is unavailable.
- [x] 4.3 Run the web service test suite and confirm the new degraded-mode behavior matches the updated specs.
