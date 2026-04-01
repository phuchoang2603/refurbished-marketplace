# Refurbished Marketplace Plan

## Current Status

- Repository bootstrapped as a single Go module with `go.mod` and `go.sum`.
- Initial services exist: `users`, `products`, `orders`.
- Architecture direction is now explicit: REST at edge (`web` service), gRPC for all internal service-to-service traffic.
- Development now standardizes on Kubernetes (Tilt + Helm + CloudNativePG).
- `users` is the first implemented vertical slice (migration + sqlc queries + service + handlers + integration tests).
- Users auth is implemented with JWT login/refresh/logout and DB-backed refresh token sessions.
- Web edge service exists and now owns REST entrypoints while users is served via gRPC.
- `orders` vertical slice is implemented as gRPC-first with PostgreSQL migrations, sqlc, and service tests.

## Data Model Direction

| Service  | Responsibility       | Key Fields Added                                      |
| -------- | -------------------- | ----------------------------------------------------- |
| Users    | Identity & Profile   | `x_pos`, `y_pos`                                      |
| Products | Inventory & Merchant | `terminal_id`, `x_pos`, `y_pos`                       |
| Orders   | Intent to Buy        | `status` (`PENDING`, `PAID`, `FAILED`), `total_cents` |
| Payment  | Financial & ML Logic | `tx_fraud`, `tx_fraud_scenario`, `tx_time_seconds`    |

## Schema Next Step

Implement the schema changes above in the database layer first, then update the service code around those schemas.

- `users`: add location columns only; derive spending aggregates from transaction history when needed.
- `products`: add merchant/terminal metadata columns.
- `orders`: keep order header state and totals.
- `payment`: add fraud/transaction tracking fields when the service is introduced.

## Canonical Repo Tree

```text
/
  PLAN.md
  SPEC.MD
  flake.nix
  go.mod
  go.sum
  Makefile
  Tiltfile
  services/
    web/
      cmd/web/
      internal/
        handlers/
        usersclient/
      tests/
    users/
      cmd/users/
      db/migrations/
      db/queries/
      internal/
      tests/
    products/
      cmd/products/
      db/migrations/
      db/queries/
      internal/
      tests/
    orders/
      cmd/orders/
      db/migrations/
      db/queries/
      internal/
      tests/
  shared/
    contracts/
    messaging/
    proto/
      users/v1/
      usersclient/
    tracing/
    testutil/
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

## Users Service (Implemented)

- Runtime:
  - `services/users/cmd/users/main.go`
  - requires `DB_URL` (no hardcoded fallback)
  - serves gRPC on `GRPC_ADDR` (default `:9091`)
- SQL and migrations:
  - migrations:
    - `001_users.sql`
    - `002_refresh_tokens.sql`
  - queries:
    - user queries: `CreateUser`, `GetUserByID`, `GetUserByEmail`
    - auth session queries: `CreateRefreshToken`, `GetRefreshTokenByID`, `RevokeRefreshToken`
- sqlc:
  - config at `services/users/sqlc.yaml`
  - generated package at `services/users/internal/database`
- Service layer:
  - validation + password hashing + query orchestration in `internal/service`
- Auth endpoints:
  - `POST /auth/login`
  - `POST /auth/refresh`
  - `POST /auth/logout`
- Auth config:
  - required env: `JWT_SECRET`
  - code defaults:
    - issuer: `refurbished-marketplace`
    - audience: `refurbished-marketplace-api`
    - access TTL: `15m`
    - refresh TTL: `168h`

## Focus Areas

- Keep `users` as identity/profile plus auth session source of truth.
- Keep `products` as read-only catalog data in the public API.
- Keep `orders` as order headers plus line items.
- Introduce `payment` as the bank/fraud boundary later.
- Update docs in `docs/architecture/` when a topic becomes stable enough to move out of `PLAN.md`.

## Minimal Schema

- `users`: `id`, `email`, `password_hash`, `x_pos`, `y_pos`
- `products`: `id`, `name`, `description`, `price_cents`, `stock`, `terminal_id`, `x_pos`, `y_pos`
- `orders`: `id`, `buyer_user_id`, `status`, `total_cents`
- `order_items`: `id`, `order_id`, `product_id`, `quantity`, `unit_price_cents`, `line_total_cents`
- `payment`: `id`, `order_id`, `tx_fraud`, `tx_fraud_scenario`, `tx_time_seconds`

## Next Steps

1. Implement the DB schema changes for `users`, `products`, `orders`, and the future `payment` boundary.
2. Remove remaining product write paths from the public web API.
3. Add the `payment` service and its bank/fraud event flow.
4. Introduce Kafka/Strimzi contracts for payment outcomes and downstream consumers.
