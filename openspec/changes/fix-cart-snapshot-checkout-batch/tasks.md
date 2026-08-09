## 1. Contracts (proto)

- [ ] 1.1 Extend `CartItem` with `product_name`, `unit_price_cents`
- [ ] 1.2 Extend `AddCartItemRequest` / `SetCartItemQuantityRequest` with the same snapshot fields
- [ ] 1.3 Add `RemoveCartItems` RPC + request/response messages
- [ ] 1.4 Add products `GetProductsByIDs` RPC + messages
- [ ] 1.5 Run `generate-proto` (and tidy modules if required)

## 2. Cart service

- [ ] 2.1 Extend `service.CartItem` and JSON persistence to carry snapshot fields
- [ ] 2.2 `AddCartItem` / `SetCartItemQuantity` accept and store snapshots (refresh on existing line)
- [ ] 2.3 Reject Add/Set without non-empty `product_name` or with non-positive `unit_price_cents`
- [ ] 2.4 Implement `RemoveCartItems` (one load, drop set of product_ids, one save; ignore unknown ids; empty ids → InvalidArgument)
- [ ] 2.5 Map new fields in gRPC server; wire `RemoveCartItems` handler
- [ ] 2.6 Update cart unit tests that assert item shape / remove behavior / snapshot validation

## 3. Products service

- [ ] 3.1 Add sqlc query `GetProductsByIDs` (`id = ANY(...)` with inventory join aligned to `GetProductByID`)
- [ ] 3.2 Implement service method with max 100 IDs guard
- [ ] 3.3 Map gRPC handler to existing product row mapper

## 4. Web BFF

- [ ] 4.1 Products client + `Dependencies` / fakes: `GetProductsByIDs`
- [ ] 4.2 Cart client + deps/fakes: snapshot args on Add/Set; `RemoveCartItems`
- [ ] 4.3 `handleAddCartItem`: GetProductByID → merchant match → Add with stamp
- [ ] 4.4 `handleSetCartItemQuantity`: same stamp path when quantity > 0
- [ ] 4.5 `mapCartView`: snapshots only; incomplete lines as not available
- [ ] 4.6 `buildCheckoutOrderItems`: one `GetProductsByIDs`; SoR prices; fail closed
- [ ] 4.7 `removeCheckedOutItems`: single `RemoveCartItems` (+ cookie clear if empty)

## 5. Verification checklist (when implementing)

- [ ] 5.1 Add 2 products → Redis/GetCart includes names + prices; `/cart` works without per-id product GETs
- [ ] 5.2 Change catalog price → cart may show old; checkout uses new price on order
- [ ] 5.3 Multi-item merchant checkout → one products batch + one multi-remove
- [ ] 5.4 Single-item UI remove still works
- [ ] 5.5 Empty cart cookie after removing last lines still works
