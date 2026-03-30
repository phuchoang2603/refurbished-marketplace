# Products Vertical Slice Plan

## Goal

Implement the first products vertical slice with the same conventions as users:

- Goose migrations
- SQLC queries
- service layer
- gRPC handlers
- tests in `services/products/tests/`

## Scope

- Service: `services/products`
- Transport: gRPC (internal service)
- Storage: PostgreSQL (`products_db`)
- Infra: existing Tilt + Helm + CloudNativePG and Compose setup

## Initial Product Model

Use a minimal model for v1:

- `id UUID PRIMARY KEY`
- `seller_user_id UUID NOT NULL`
- `name TEXT NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `condition_grade TEXT NOT NULL`
- `price_cents BIGINT NOT NULL`
- `currency TEXT NOT NULL`
- `quantity INT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

## gRPC Methods (v1)

- `CreateProduct`
- `GetProductByID`
- `ListProducts`

## SQLC Queries (v1)

- `CreateProduct`
- `GetProductByID`
- `ListProducts`

## Validation Rules

- `name`: non-empty
- `condition_grade`: non-empty for now
- `price_cents`: must be positive
- `currency`: ISO-like uppercase (start with strict 3 chars)
- `quantity`: zero or positive

## Testing

All tests stay under `services/products/tests/`:

- integration tests for create/get/list and constraints using `shared/testutil`
- service tests for validation/error mapping

## Implementation Sequence

1. Add migration `001_products.sql`.
2. Add SQL queries and run `sqlc generate`.
3. Add `internal/service` methods and error types.
4. Add protobuf contract in `services/products/proto/v1/` and generate code.
5. Add gRPC handlers/server wiring.
5. Wire `cmd/products/main.go` with required `DB_URL`.
6. Add tests in `services/products/tests/`.
7. Add products migrator Dockerfile and enable migration job in k8s chart values.
