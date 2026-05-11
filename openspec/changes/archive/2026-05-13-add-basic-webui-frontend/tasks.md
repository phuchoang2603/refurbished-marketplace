## 1. Static UI Foundation

- [x] 1.1 Add static asset serving in `services/web`, and update `infra/docker/web.Dockerfile` to copy `services/web/static` into `/static` in the final image.
- [x] 1.2 Copy DatastarUI-inspired `templ` primitives into `services/web/internal/views/` and use them as broadly as practical.
- [x] 1.3 Add `datastar-go` dependency for Datastar SSE and patch response helpers.
- [x] 1.4 Create a shared stylesheet for layout, navigation, cards, forms, tables, buttons, and responsive behavior.
- [x] 1.5 Update `AppShell` to reference static CSS, keep Datastar loading, use DatastarUI-compatible primitives where available, and remove most inline styling.
- [x] 1.6 Regenerate `templ` output after shell/static asset changes.

## 2. Marketplace Page UI

- [x] 2.1 Improve catalog list markup with clearer product cards or table layout and responsive behavior.
- [x] 2.2 Improve product detail markup with clearer pricing, description, metadata, and add-to-cart placement.
- [x] 2.3 Compose cart item product details through the products gRPC client in web handlers/view models.
- [x] 2.4 Improve cart markup with product names, current prices, estimated totals, quantity controls, remove actions, empty state, and checkout action.
- [x] 2.5 Improve order list/detail markup with clearer status, totals, item rows, and empty state.
- [x] 2.6 Improve user/account and message pages so errors and confirmations use consistent UI treatment.

## 3. Browser Auth Cookies

- [x] 3.1 Add web auth cookie names, cookie options, and helpers for setting, reading, and clearing auth cookies.
- [x] 3.2 Update login handling to set HTTP-only auth cookies using the users service token expiry values directly.
- [x] 3.3 Update logout handling to revoke through users service and clear browser auth cookies.
- [x] 3.4 Update protected browser auth middleware to accept access tokens from cookies as well as existing bearer headers where needed.

## 4. Datastar Interaction Refinement

- [x] 4.1 Ensure add-to-cart responses target an element present on product detail pages.
- [x] 4.2 Use `datastar-go` SSE patch helpers for interactive endpoints that update multiple targets, signals, or redirects.
- [x] 4.3 Ensure cart update and remove actions preserve stable DOM IDs for Datastar patching.
- [x] 4.4 Ensure auth forms and invalid submissions return browser-friendly HTML responses.
- [x] 4.5 Keep form submissions HTML/form-first without adding browser-facing JSON DTOs.

## 5. Tests And Verification

- [x] 5.1 Update `services/web/tests` to verify shell CSS references and static asset serving.
- [x] 5.2 Add or update tests for DatastarUI/static primitive usage where stable enough to assert.
- [x] 5.3 Add or update tests for catalog, product detail, composed cart rows, auth cookies, and order UI markup.
- [x] 5.4 Add or update tests for expected Datastar attributes, SDK-backed response behavior where practical, and fragment target IDs.
- [x] 5.5 Run `templ generate` and `go test ./...` in `services/web`.
