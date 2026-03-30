# Refurbished Marketplace Plan

## Current Status

- Repository bootstrapped as a single Go module with `go.mod` and `go.sum`.
- Initial services exist: `users`, `products`, `orders`.
- Development supports both Kubernetes (Tilt + Helm + CloudNativePG) and Docker Compose.
- `users` is the first implemented vertical slice (migration + sqlc queries + service + handlers + integration tests).

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
    users/
      cmd/users/
      db/migrations/
      db/queries/
      internal/
      proto/v1/
      tests/
    products/
      cmd/products/
      db/migrations/
      db/queries/
      internal/
      proto/v1/
      tests/
    orders/
      cmd/orders/
      db/migrations/
      db/queries/
      internal/
      proto/v1/
      tests/
  shared/
    contracts/
    messaging/
    proto/
    tracing/
    testutil/
  infra/
    development/
      docker/
      helm/refurbished-marketplace/
      k8s/
    production/
      docker/
      k8s/
  docs/
  web/
```

## Stack and Conventions

- Language: Go, standard library first.
- Communication: REST externally first, gRPC incrementally for internal service-to-service calls.
- Database: PostgreSQL.
- Migrations: `goose`.
- Query generation: `sqlc`.
- Event bus target: RabbitMQ (not wired yet).
- Style: small packages, explicit SQL, straightforward handlers, table-driven tests.

## Service Layout Rules

- Start simple per service; avoid over-abstracting.
- Keep service code in `services/<name>/` and private code in `internal/`.
- Keep SQL and migrations service-local:
  - `services/<service>/db/migrations/`
  - `services/<service>/db/queries/`
- Keep proto service-local first: `services/<service>/proto/v1/`.
- Keep all service tests in `services/<service>/tests/` (unit + integration).

## Users Service (Implemented)

- Runtime:
  - `services/users/cmd/users/main.go`
  - requires `DB_URL` (no hardcoded fallback)
  - serves `GET /healthz`, `POST /users`, `GET /users/{id}` via standard `net/http`
- SQL and migrations:
  - migration `001_users.sql`
  - queries: `CreateUser`, `GetUserByID`, `GetUserByEmail`
- sqlc:
  - config at `services/users/sqlc.yaml`
  - generated package at `services/users/internal/database`
- Service layer:
  - validation + password hashing + query orchestration in `internal/service`

## Auth Plan (Next)

- Keep auth business logic in `services/users` for now (not `shared/` yet).
- Add login and refresh endpoints in users service.
- Store refresh tokens in DB with a dedicated table and revocation support.
- Prefer storing token hashes (or token id + hash), not raw token text, when finalizing schema.
- Istio can validate JWT at ingress/mesh policy level, but token issuance, refresh flow, and session revocation remain app responsibilities.
- Use one shared JWT secret for both access and refresh tokens initially (`JWT_SECRET`) for simplicity.
- Keep issuer/audience/TTL defaults in code (no required env vars):
  - issuer: `refurbished-marketplace`
  - audience: `refurbished-marketplace-api`
  - access TTL: `15m`
  - refresh TTL: `168h`
- Keep secret externalized via env/secret manager (do not hardcode `JWT_SECRET` in code).

## Kubernetes Development (Current)

- Orchestration:
  - `Tiltfile` uses Helm chart under `infra/development/helm/refurbished-marketplace`
  - CloudNativePG operator installed through Helm in namespace `cnpg-system`
- Namespaces:
  - each service/database pair isolated in its own namespace (`users`, `products`, `orders`)
- Database readiness:
  - service deployments use `initContainer` with `pg_isready` before app container startup
- Migration jobs:
  - Helm hook jobs (`pre-install,pre-upgrade`) in `templates/migrations.tpl`
  - users migration enabled by default
  - users migrator image built from `infra/development/docker/users-migrator.Dockerfile` (base: `ghcr.io/kukymbr/goose-docker:3.27.0`)

## Docker Compose Development (Current)

- File: `infra/development/docker/compose.yaml`
- Runs `users`, `products`, `orders` + dedicated DBs (`users-db`, `products-db`, `orders-db`).
- `DB_URL` provided per app service.
- Compose config is DRY using YAML anchors for shared DB settings.

## Testing Strategy (Current)

- Test location:
  - keep all service tests in `services/<service>/tests/`
- Users tests:
  - `services/users/tests/integration_test.go` uses Testcontainers PostgreSQL module + Goose migrations
  - validates create/read, missing-user behavior, and unique-email constraint behavior
- Shared test utilities:
  - `shared/testutil/postgres.go` contains reusable Postgres+Goose setup logic for future service tests

## Environment and Tooling

- Nix + direnv enabled.
- `.envrc` supports Colima/Testcontainers compatibility.
- `Makefile` includes:
  - `generate-proto`

## Next Steps

1. Implement `products` vertical slice (`goose` + `sqlc` + handlers + tests) mirroring users.
2. Implement `orders` vertical slice with DB and API surface.
3. Implement users auth (JWT access + refresh, DB-backed refresh sessions, revocation/rotation).
4. Add first gRPC contract for internal lookup flows (service-local proto).
5. Introduce RabbitMQ contracts and one async workflow.
6. Add migration jobs/migrator images for products and orders once migrations exist.
