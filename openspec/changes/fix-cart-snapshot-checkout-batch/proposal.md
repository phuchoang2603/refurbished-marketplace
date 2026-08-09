## Why

Cart and merchant-scoped checkout are chatty at the web BFF:

1. **Cart UI composition** — `mapCartView` does `1 + N` RPCs (`GetCart` then per-line `Products.GetProductByID`). Quantity/remove fragments re-run the same loop.
2. **Checkout composition** — `buildCheckoutOrderItems` again does **N** `GetProductByID` for money-path prices; then `removeCheckedOutItems` does **N** sequential `RemoveCartItem` Redis read-modify-writes.

Catalog Meilisearch CQRS (#7) does **not** fix known-ID cart hydration. Issue [#33](https://github.com/phuchoang2603/refurbished-marketplace/issues/33) covers **replication on the cart line**, **one batch Products re-validation at checkout**, and **multi-remove**.

## What Changes

- **Cart line snapshot** (ephemeral Redis JSON / cart proto): persist display facts with each item — at least `product_name` + `unit_price_cents` (alongside existing `product_id`, `merchant_id`, `quantity`)
- **BREAKING (internal cart gRPC):** `AddCartItem` / `SetCartItemQuantity` require caller-supplied snapshot fields (non-empty name, positive unit price); sole expected client is web
- **Stamp snapshots at write time in web** after a single-SKU product read on Add/Set; cart service remains free of products dependency
- **Cart page / cart fragments** render only from cart payloads (no per-line Products loop)
- **Products `GetProductsByIDs`** batch RPC — Postgres authoritative, used for checkout re-validation only (not for casual cart paint)
- **Cart `RemoveCartItems`** multi-id RPC — one Redis load/filter/save; checkout removes the merchant group’s product IDs in one call
- Keep single-item `RemoveCartItem` for the UI remove button (may wrap multi internally)

### Non-goals

- Cart service calling Products / Meilisearch / any catalog client
- Using Meilisearch (or other read models) as cart display source
- Live stock / availability badges on the cart page from Products
- User-facing “price changed since you added” toast/dialog in v1 (fail closed on missing product / wrong merchant; price is always SoR at checkout)
- Multi-product “add all” bulk API
- Migrating/rewriting in-flight Redis carts with a background job (legacy lines without snapshot fields render as incomplete / not available until next Add that re-stamps, or user removes)
- Changing order/payment reservation semantics

## Capabilities

### New Capabilities

_(none — extends existing cart / products / web)_

### Modified Capabilities

- `cart`: cart items carry product display snapshots; multi-remove RPC; add/set accept snapshot fields
- `products`: batch `GetProductsByIDs` for authoritative multi-get by ID
- `web`: stamp snapshots on cart mutations; map cart view from snapshots; checkout uses batch product fetch + multi-remove

## Impact

- **Protos:** `shared/proto/cart/v1`, `shared/proto/products/v1` + `generate-proto`
- **Services:** `services/cart` (storage shape, remove-many), `services/products` (SQL/sqlc + gRPC), `services/web` (handlers, clients, deps/fakes, views if needed)
- **Issue:** Implements [#33](https://github.com/phuchoang2603/refurbished-marketplace/issues/33)
- **Orthogonal:** [#7](https://github.com/phuchoang2603/refurbished-marketplace/issues/7) / `add-meilisearch-catalog-read-model`
