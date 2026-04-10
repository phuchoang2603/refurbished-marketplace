# Testing Strategy

This repository uses a pragmatic testing policy focused on service confidence without duplication.

## Principles

- Prefer service-level integration tests as the primary confidence layer.
- Add pure unit tests only for non-trivial, pure logic that does not need database setup.

## Running Tests (Current Repo Layout)

This repo is **multi-module** (each `services/<service>` and `shared/` has its own `go.mod`), so the most reliable way to run tests is from within a module directory:

- `cd shared && go test ./...`
- `cd services/users && go test ./...`
- `cd services/products && go test ./...`
- `cd services/orders && go test ./...`
- `cd services/cart && go test ./...`
- `cd services/inventory && go test ./...`

The root `Makefile` currently runs `go test ./...` from repo root; that will only work if you’ve set up a Go workspace (for example via `go work`) that includes the modules.

## What to Test

- **Service + DB integration tests (primary):**
  - run against temporary Postgres using Testcontainers and Goose migrations
  - cover business rules, data persistence, and error behavior
- **Unit tests (selective):**
  - validation helpers
  - token or parsing logic
  - deterministic mappers and pagination normalization

## What to Avoid

- Duplicating the same scenario across multiple test layers without new coverage value.
- Building large HTTP + gRPC + DB end-to-end chains for every feature unless required.

## Current Scope

- Test location:
  - keep all service tests in `services/<service>/tests/`
- Users tests:
  - `services/users/tests/service_test.go` validates auth/login/refresh/logout and user service behavior
  - coverage includes create/read, missing-user behavior, unique-email constraint, refresh rotation, and logout revocation
- Products tests:
  - `services/products/tests/service_test.go` validates product service create/read/list behavior and query-level no-row behavior
- Orders tests:
  - `services/orders/tests/service_test.go` validates order service create/read/list/state transition behavior
- Shared test utilities:
  - `shared/testutil/postgres.go` contains reusable Postgres+Goose setup logic for future service tests

## Scope Direction

- Users schema should add profile/location/spending fields.
- Products schema should add terminal/location fields.
- Orders should keep transaction-safe order headers and line items.
- Payment testing will be added when the service exists.
