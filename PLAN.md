# Refurbished Marketplace Plan

## Current Status

- Repository bootstrapped as a single Go module with `go.mod` and `go.sum`.
- Initial services exist: `users`, `products`, `orders`, `inventory`.
- Architecture direction is now explicit: REST at edge (`web` service), gRPC for all internal service-to-service traffic.
- Development now standardizes on Kubernetes (Tilt + Helm + CloudNativePG).
- `users` is the first implemented vertical slice (migration + sqlc queries + service + handlers + integration tests).
- Users auth is implemented with JWT login/refresh/logout and DB-backed refresh token sessions.
- Web edge service exists and now owns REST entrypoints while users is served via gRPC.
- `orders` vertical slice is implemented as gRPC-first with PostgreSQL migrations, sqlc, service tests, and a transactional outbox row per `orders.item.created` event.
- `products` is catalog-only now; stock moved out into `inventory`.
- `inventory` is scaffolded as the stock/reservation service.

## Canonical Repo Tree

```text
/
  PLAN.md
  SPEC.MD
  flake.nix
  Makefile
  Tiltfile
  services/
    web/
      cmd/web/
      internal/
        handlers/
        usersclient/
      tests/
      go.mod
      go.sum
    users/
      cmd/users/
      db/migrations/
      db/queries/
      internal/
      tests/
      go.mod
      go.sum
    products/
      cmd/products/
      db/migrations/
      db/queries/
      internal/
      tests/
      go.mod
      go.sum
    orders/
      cmd/orders/
      db/migrations/
      db/queries/
      internal/
      tests/
      go.mod
      go.sum
  shared/
    messaging/
    proto/
      users/v1/
      usersclient/
    tracing/
    testutil/
    go.mod
    go.sum
  infra/
    charts/
      refurbished-marketplace/
        templates/
        Chart.yaml
    development/
      docker/
      k8s/
        dev-helm-values.yaml
        secrets.yaml
    production/
      docker/
      k8s/
  docs/
  web/
```

## Stack and Conventions

- Language: Go, standard library first.
- Communication: REST only at edge/web layer, gRPC inside the microservice mesh.
- Database: PostgreSQL.
- Migrations: `goose`.
- Query generation: `sqlc`.
- Event bus target: Kafka via Strimzi (preferred for future recommender/ML integrations; not wired yet).
- Style: small packages, explicit SQL, straightforward handlers, table-driven tests.

## Service Layout Rules

- Start simple per service; avoid over-abstracting.
- Keep service code in `services/<name>/` and private code in `internal/`.
- Internal services should expose gRPC contracts first (`proto/v1`) and avoid new REST handlers.
- Shared gRPC contracts live under `shared/proto/<domain>/v1/` when reused by multiple services.
- Keep REST/HTTP DTO shaping in the web/edge service.
- Keep SQL and migrations service-local:
  - `services/<service>/db/migrations/`
  - `services/<service>/db/queries/`
- Keep all service tests in `services/<service>/tests/` (unit + integration).

## Focus Areas

- Keep `users` as identity/profile plus auth session source of truth.
- Keep `products` as catalog data with internal/admin writes.
- Keep `inventory` as the source of truth for available/reserved stock.
- Keep `orders` as order headers plus line items.
- Keep `cart` separate from order state and payment state.
- Introduce `payment` as the bank/fraud boundary later.

## Eventing Reliability

- Kafka remains the async backbone for downstream consumers.
- `orders` and later `payment` should persist domain events to a local outbox table inside the same database transaction as the business write.
- Debezium should stream outbox rows from Postgres into Kafka.
- Consumers such as `payment` should use an inbox table to dedupe repeated deliveries.
- Fraud and analytics should consume the canonical event stream, not application-generated ad hoc payloads.

## Minimal Schema

- `users`: `id`, `email`, `password_hash`, `x_pos`, `y_pos`
- `products`: `id`, `name`, `description`, `price_cents`, `merchant_id`, `terminal_id`, `x_pos`, `y_pos`
- `inventory`: `product_id`, `available_qty`, `reserved_qty`
- `orders`: `id`, `buyer_user_id`, `status`, `total_cents`
- `order_items`: `id`, `order_id`, `product_id`, `merchant_id`, `quantity`, `unit_price_cents`, `line_total_cents`
- `orders_outbox`: `id`, `aggregate_id`, `event_type`, `payload`, `publish_attempts`, `created_at`, `published_at`
- `cart`: Redis session state only; no Postgres schema required
- `payment`: `id`, `order_id`, `merchant_id`, `tx_fraud`, `tx_fraud_scenario`, `tx_time_seconds`

## Next Steps

1. Implement inventory consumption of `orders.item.created` keyed by `product_id`.
2. Add merchant snapshots to `products`, `order_items`, and `payment` flows.
3. Add the outbox publisher/CDC path from `orders` to Kafka with `product_id` message keys.
4. Add the `payment` service and its bank/fraud event flow.
5. Introduce inbox dedupe in consumers that need at-least-once safety.
6. Add admin orchestration later if you want a single product + inventory creation path.
