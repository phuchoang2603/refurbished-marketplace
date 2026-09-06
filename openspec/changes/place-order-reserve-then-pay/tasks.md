## 1. Contracts

- [x] 1.1 Add `idempotency_key` to the orders create/place-order protobuf under `shared/proto/orders/v1/` and regenerate with devenv `generate-proto`
- [x] 1.2 Add products `ReserveStock` (order id + lines) protobuf under `shared/proto/products/v1/` and regenerate

## 2. Products reserve command

- [x] 2.1 Expose `ReserveStock` gRPC in `services/products/internal/grpcserver/` that shares reserve transactional logic with `HandleOrdersCreated` (`reservation_helper.go`)
- [x] 2.2 Emit `inventory.reserved` / `inventory.reservation_failed` outbox from the command path the same way as the Kafka path
- [x] 2.3 Make `HandleOrdersCreated` a no-op hold when the order already has reservations (no double `available_qty` decrement)

## 3. Orders PlaceOrder

- [x] 3.1 Goose migration + sqlc: persist `idempotency_key` with unique `(buyer_user_id, idempotency_key)` under `services/orders/db/`
- [x] 3.2 Add a products gRPC client to orders (`PRODUCTS_SVC_ADDR` in `infra/charts/refurbished-marketplace/values.yaml`)
- [x] 3.3 Place-order: insert order + items + `orders.created` outbox, then `ReserveStock`; same key returns existing order; conflicting body is an error; reserve failure must not leave a payable unreserved order
- [x] 3.4 Allow `orders` → `products` gRPC in `infra/charts/refurbished-marketplace/templates/mesh-policy.tpl` (today only `web` may call gRPC services)

## 4. Payment session reattach

- [x] 4.1 `CreateHostedPaymentSession` in `services/payment`: get-or-create by `order_id`; refresh or replace expired unpaid sessions so a new hosted URL is returned for the same order

## 5. Web edge

- [x] 5.1 Checkout form hidden intent UUID reused on resubmit (`services/web/internal/views/cart/` + checkout handler)
- [x] 5.2 Checkout: batch re-validate → PlaceOrder with key → session create → 303; remove cart multi-remove from this POST (`checkout.go`)
- [x] 5.3 After successful hosted-payment callback, multi-remove paid product IDs when `cart_id` is present
- [x] 5.4 Resume-payment action on the unpaid order page (get-or-create session + redirect)

## 6. Docs

- [x] 6.1 Update `docs/order-placement.md` so reserve happens on PlaceOrder gRPC before hosted redirect, and cart drain is after paid
