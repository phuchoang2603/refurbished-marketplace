# Refurbished Marketplace Plan

## Current Status

- Repository bootstrapped as a single Go module with `go.mod` and `go.sum`.
- Initial services exist: `users`, `products`, `orders`.
- Architecture direction is now explicit: REST at edge (`web` service), gRPC for all internal service-to-service traffic.
- Development now standardizes on Kubernetes (Tilt + Helm + CloudNativePG).
- `users` is the first implemented vertical slice (migration + sqlc queries + service + handlers + integration tests).
- Users auth is implemented with JWT login/refresh/logout and DB-backed refresh token sessions.
- Web edge service exists and now owns REST entrypoints while users is served via gRPC.

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
- Event bus target: RabbitMQ (not wired yet).
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

## Transport Strategy (Committed)

- Edge:
  - `web` service owns client-facing REST APIs.
  - Web service composes calls to internal services via gRPC clients.
- Internal:
  - `users`, `products`, and `orders` communicate through gRPC only.
  - Domain events continue to use RabbitMQ for async workflows.
- Transition:
  - Existing users REST endpoints are considered transitional and will be replaced by web-edge REST + users gRPC.

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

## Kubernetes Development (Current)

- Orchestration:
  - `Tiltfile` uses Helm chart under `infra/charts/refurbished-marketplace`
  - development values are in `infra/development/k8s/dev-helm-values.yaml`
  - Helm release namespace is `ecommerce`
  - CloudNativePG operator installed through Helm in namespace `cnpg-system`
- Secrets:
  - secrets are applied separately via `infra/development/k8s/secrets.yaml`
  - service and migration manifests consume secrets using `secretKeyRef`
- Namespaces:
  - application resources (`web`, `users`, `products`, `orders`, db clusters, migration jobs) are deployed to `ecommerce`
  - CloudNativePG operator stays in `cnpg-system`
- Database readiness:
  - service deployments use `initContainer` with `pg_isready` before app container startup
- Migration jobs:
  - Helm hook jobs (`pre-install,pre-upgrade`) in `templates/migrations.tpl`
  - users migration enabled by default
  - users migrator image built from `infra/development/docker/users-migrator.Dockerfile` (base: `ghcr.io/kukymbr/goose-docker:3.27.0`)
  - products migration enabled by default
  - products migrator image built from `infra/development/docker/products-migrator.Dockerfile` (base: `ghcr.io/kukymbr/goose-docker:3.27.0`)
- Service ports:
  - web: `8080`
  - users gRPC: `9091`
  - products gRPC: `9092`
  - users-db: 5432 (Postgres)
  - products-db: 5433 (Postgres)
  - orders-db: 5434 (Postgres)

## Service Discovery (Current)

- Web service upstream addresses are defined in values under `services.web.env`.
- Current values include:
  - `USERS_SVC_ADDR=users:9091`
  - `PRODUCTS_SVC_ADDR=products:9092`
  - `ORDERS_SVC_ADDR=orders:8083`

## gRPC Contracts and Clients (Current)

- Users protobuf contract is centralized at `shared/proto/users/v1/users.proto`.
- Generated users gRPC code lives in `shared/proto/users/v1/`.
- Reusable users gRPC client lives in `shared/proto/usersclient/`.
- Products protobuf contract is centralized at `shared/proto/products/v1/products.proto`.
- Generated products gRPC code lives in `shared/proto/products/v1/`.
- Reusable products gRPC client lives in `shared/proto/productsclient/`.

## Testing Strategy (Current)

- Test location:
  - keep all service tests in `services/<service>/tests/`
- Users tests:
  - `services/users/tests/service_test.go` validates auth/login/refresh/logout and user service behavior
  - coverage includes create/read, missing-user behavior, unique-email constraint, refresh rotation, and logout revocation
- Products tests:
  - `services/products/tests/service_test.go` validates product service create/read/list behavior and query-level no-row behavior
- Shared test utilities:
  - `shared/testutil/postgres.go` contains reusable Postgres+Goose setup logic for future service tests

## Products Service (In Progress)

- Runtime:
  - `services/products/cmd/products/main.go`
  - requires `DB_URL` (no hardcoded fallback)
  - serves gRPC on `GRPC_ADDR` (default `:9092`)
- SQL and migrations:
  - migrations:
    - `001_products.sql`
  - queries:
    - `CreateProduct`, `GetProductByID`, `ListProducts`
- sqlc:
  - config at `services/products/sqlc.yaml`
  - generated package at `services/products/internal/database`
- Service layer:
  - product validation + create/read/list logic in `services/products/internal/service`
- gRPC server:
  - `services/products/internal/grpcserver`
- Web edge endpoints:
  - `POST /products`
  - `GET /products/{id}`
  - `GET /products?limit=&offset=`
  - `PATCH /products/{id}` (planned)
  - `DELETE /products/{id}` (planned)
- Tests:
  - `services/products/tests/service_test.go` covers service/query behavior

## Auth + Authorization Direction (Committed)

- Users service remains source of truth for auth session domain:
  - login/refresh/logout token issuance and refresh session state
- Web service acts as edge authorization gate:
  - validate access JWT for protected REST endpoints
  - extract authenticated user id (`sub`) and pass trusted identity to internal services
- Products ownership model:
  - each product must store `owner_user_id`
  - create/update/delete must require authenticated user
  - list/get stay public for marketplace browsing
- Planned protected REST endpoints:
  - `POST /products`
  - `PATCH /products/{id}`
  - `DELETE /products/{id}`

## Environment and Tooling

- Nix + direnv enabled.
- `.envrc` supports Colima/Testcontainers compatibility.
- `Makefile` includes:
  - `generate-proto`

## Next Steps

1. Add web JWT middleware and endpoint protection for product mutations.
2. Extend products model/queries/proto with `owner_user_id` and mutation methods (`UpdateProduct`, `DeleteProduct`).
3. Implement `PATCH /products/{id}` and `DELETE /products/{id}` in web and enforce owner-only authorization.
4. Implement `orders` vertical slice with gRPC-first transport (`goose` + `sqlc` + service + tests).
5. Add orders migration job + orders migrator image once orders migrations exist.
6. Introduce RabbitMQ contracts and one async workflow.
