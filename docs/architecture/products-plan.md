# Products Vertical Slice Plan

## Goal

Implement the first products vertical slice with the same conventions as users:

- Goose migrations
- SQLC queries
- service layer
- gRPC handlers
- tests in `services/products/tests/`
- normal ecommerce product lifecycle (not C2C inspection/escrow)
- keep stock management out of products; inventory owns reservations

## Scope

- Service: `services/products`
- Transport: gRPC (internal service)
- Storage: PostgreSQL (`products_db`)
- Infra: existing Tilt + Helm + CloudNativePG and Compose setup

## Product Model

Use a minimal catalog model for v1:

- `id UUID PRIMARY KEY`
- `terminal_id UUID NOT NULL`
- `x_pos DOUBLE PRECISION NOT NULL`
- `y_pos DOUBLE PRECISION NOT NULL`
- `name TEXT NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `price_cents BIGINT NOT NULL`
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
- `price_cents`: must be positive
- `terminal_id`: must be valid UUID
- `x_pos` / `y_pos`: required coordinate values

## Testing

All tests stay under `services/products/tests/`:

- integration tests for create/get/list and constraints using `shared/testutil`
- service tests for validation/error mapping
- no inventory mutation tests here; those belong to `services/inventory/tests/`

## Implementation Sequence

1. Add migration `001_products.sql`.
2. Add SQL queries and run `sqlc generate`.
3. Add `internal/service` methods and error types.
4. Add protobuf contract in `shared/proto/products/v1/` and generate code.
5. Add gRPC handlers/server wiring.
6. Wire `cmd/products/main.go` with required `DB_URL`.
7. Add tests in `services/products/tests/`.
8. Add products migrator Dockerfile and enable migration job in k8s chart values.
9. Keep product writes out of the public web API.
