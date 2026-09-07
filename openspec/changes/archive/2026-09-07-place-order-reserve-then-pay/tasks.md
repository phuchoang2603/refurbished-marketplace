## 1. Contracts

- [x] 1.1 Add `idempotency_key` to the orders create/place-order protobuf under `shared/proto/orders/v1/` and regenerate with devenv `generate-proto`
- [x] 1.2 Add products `ReserveStock` (order id + lines) protobuf under `shared/proto/products/v1/` and regenerate

## 2. Products reserve command

- [x] 2.1 Expose `ReserveStock` gRPC in `services/products/internal/grpcserver/` that shares reserve transactional logic with `HandleOrdersCreated` (`reservation_helper.go`)
- [x] 2.2 Emit `inventory.reserved` / `inventory.reservation_failed` outbox from the command path the same way as the Kafka path
- [x] 2.3 Make `HandleOrdersCreated` a no-op hold when the order already has reservations (no double `available_qty` decrement)

## 3. Orders PlaceOrder

- [x] 3.1 Goose migration + sqlc: persist `idempotency_key` with unique `(buyer_user_id, idempotency_key)` under `services/orders/db/`
- [x] 3.2 Orders persist + `orders.created` only; no products gRPC client and no `PRODUCTS_SVC_ADDR` on orders
- [x] 3.3 Same key returns existing order; conflicting body is an error; failed orders are not payable
- [x] 3.4 Mesh: web already calls products; do not allow `orders` → `products`

## 4. Payment session reattach

- [x] 4.1 `CreateHostedPaymentSession` in `services/payment`: get-or-create by `order_id`; refresh or replace expired unpaid sessions so a new hosted URL is returned for the same order

## 5. Web edge

- [x] 5.1 Checkout form hidden intent UUID reused on resubmit (`services/web/internal/views/cart/` + checkout handler)
- [x] 5.2 Checkout: batch re-validate → CreateOrder → ReserveStock → session create → 303; on reserve failure mark order FAILED and rotate intent; remove cart multi-remove from this POST (`checkout.go`)
- [x] 5.3 After successful hosted-payment callback, multi-remove paid product IDs when `cart_id` is present
- [x] 5.4 Resume-payment action on the unpaid order page (re-hold stock, get-or-create session + redirect)

## 6. Docs

- [x] 6.1 Update `docs/order-placement.md` so web reserves after CreateOrder and before hosted redirect, and cart drain is after paid
