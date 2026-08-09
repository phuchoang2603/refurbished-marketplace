## 1. Contracts (proto)

- [x] 1.1 Extend `CartItem` with `product_name`, `unit_price_cents`
- [x] 1.2 Extend `AddCartItemRequest` / `SetCartItemQuantityRequest` with the same snapshot fields
- [x] 1.3 Add `RemoveCartItems` RPC + request/response messages
- [x] 1.4 Add products `GetProductsByIDs` RPC + messages
- [x] 1.5 Run `generate-proto` (and tidy modules if required)

## 2. Cart service

- [x] 2.1 Extend `service.CartItem` and JSON persistence to carry snapshot fields
- [x] 2.2 `AddCartItem` / `SetCartItemQuantity` accept and store snapshots (refresh on existing line)
- [x] 2.3 Reject Add/Set without non-empty `product_name` or with non-positive `unit_price_cents`
- [x] 2.4 Implement `RemoveCartItems` (one load, drop set of product_ids, one save; ignore unknown ids; empty ids → InvalidArgument)
- [x] 2.5 Map new fields in gRPC server; wire `RemoveCartItems` handler
- [x] 2.6 Update cart unit tests that assert item shape / remove behavior / snapshot validation

## 3. Products service

- [x] 3.1 Add sqlc query `GetProductsByIDs` (`id = ANY(...)` with inventory join aligned to `GetProductByID`)
- [x] 3.2 Implement service method with max 100 IDs guard
- [x] 3.3 Map gRPC handler to existing product row mapper

## 4. Web BFF

- [x] 4.1 Products client + `Dependencies` / fakes: `GetProductsByIDs`
- [x] 4.2 Cart client + deps/fakes: snapshot args on Add/Set; `RemoveCartItems`
- [x] 4.3 `handleAddCartItem`: GetProductByID → merchant match → Add with stamp
- [x] 4.4 `handleSetCartItemQuantity`: same stamp path when quantity > 0
- [x] 4.5 `mapCartView`: snapshots only; incomplete lines as not available
- [x] 4.6 `buildCheckoutOrderItems`: one `GetProductsByIDs`; SoR prices; fail closed
- [x] 4.7 `removeCheckedOutItems`: single `RemoveCartItems` (+ cookie clear if empty)

## 5. Verification checklist (when implementing)

- [x] 5.1 Add 2 products → Redis/GetCart includes names + prices; `/cart` works without per-id product GETs
- [x] 5.2 Change catalog price → cart may show old; checkout uses new price on order
- [x] 5.3 Multi-item merchant checkout → one products batch + one multi-remove
- [x] 5.4 Single-item UI remove still works
- [x] 5.5 Empty cart cookie after removing last lines still works
