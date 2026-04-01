# Testing Strategy

This repository uses a pragmatic testing policy focused on service confidence without duplication.

## Principles

- Prefer service-level integration tests as the primary confidence layer.
- Add pure unit tests only for non-trivial, pure logic that does not need database setup.

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

- Users:
  - service-level integration tests in `services/users/tests/service_test.go`
- Products:
  - service-level integration tests in `services/products/tests/service_test.go`
- Orders:
  - service-level integration tests in `services/orders/tests/service_test.go`

## Scope Notes

- The repo is scoped as normal ecommerce, not C2C escrow.
- Recommender and external payment/fraud systems are later integrations, not core test targets yet.
