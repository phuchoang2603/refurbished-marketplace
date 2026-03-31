# Testing Strategy

This repository uses a pragmatic testing policy focused on service confidence without duplication.

## Principles

- Prefer service-level integration tests as the primary confidence layer.
- Keep gRPC tests as lightweight smoke tests for transport wiring and status-code mapping.
- Add pure unit tests only for non-trivial, pure logic that does not need database setup.

## What to Test

- **Service + DB integration tests (primary):**
  - run against temporary Postgres using Testcontainers and Goose migrations
  - cover business rules, data persistence, and error behavior
- **gRPC smoke tests (minimal):**
  - one happy-path RPC per service area
  - one invalid-input/error-path assertion that verifies gRPC status mapping
- **Unit tests (selective):**
  - validation helpers
  - token or parsing logic
  - deterministic mappers and pagination normalization

## What to Avoid

- Duplicating full service scenarios in both service tests and gRPC tests.
- Building large HTTP + gRPC + DB end-to-end chains for every feature unless required.

## Current Scope

- Users:
  - service-level integration tests in `services/users/tests/service_test.go`
  - gRPC smoke tests in `services/users/tests/grpc_test.go`
- Products:
  - service-level integration tests in `services/products/tests/service_test.go`
  - gRPC smoke tests in `services/products/tests/grpc_test.go`
