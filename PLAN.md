# Refurbished Marketplace Plan

## Canonical Repo Tree

```text
/
  PLAN.md
  flake.nix
  go.mod
  go.sum
  Makefile
  Tiltfile
  proto/
  services/
    api-gateway/
    users-service/
    catalog-service/
    inventory-service/
    orders-service/
    payment-service/
  shared/
    contracts/
    db/
    env/
    messaging/
    proto/
    retry/
    tracing/
    types/
    util/
  docs/
    architecture/
  infra/
    development/
      docker/
      k8s/
    production/
      docker/
      k8s/
  tools/
  web/
  assets/
```

## Service Layout Rules

- Start simple: services may begin with a flat layout when small.
- Move to `cmd/`, `internal/`, and optional `pkg/` once service complexity grows.
- Keep service business logic private under `internal/`.
- Keep transport handlers close to service code (HTTP, gRPC, event consumers).
- Each service owns its persistence and migrations.

## Proto and Codegen Plan

- Keep protobuf definitions at repository root in `proto/`.
- Version contracts by domain: `proto/<domain>/v1/*.proto`.
- Generate Go code into `shared/proto/`.
- Keep generation commands in the top-level `Makefile`.

## Shared Package Boundaries

- `shared/contracts`: HTTP, WS, and AMQP event names/payload contracts.
- `shared/env`: env parsing helpers.
- `shared/messaging`: RabbitMQ connection, publish, and consume utilities.
- `shared/tracing`: HTTP, gRPC, and RabbitMQ tracing wrappers.
- `shared/retry`: retry policy and backoff.
- `shared/db`: shared DB connection helpers.
- `shared/util` and `shared/types`: minimal cross-service helpers and types.

Only move code into `shared/` after at least two services need it.

## Data and Persistence Plan

- Primary database: PostgreSQL.
- Migrations: `goose`.
- Query layer: `sqlc`.
- Keep SQL explicit in files.
- Preferred per-service layout:
  - `services/<service>/db/migrations/`
  - `services/<service>/db/queries/`
- Keep one schema owner per service boundary.

## Messaging Plan

- Use RabbitMQ for async workflows and eventual consistency.
- Centralize routing keys/constants in `shared/contracts`.
- Use topic exchanges with clear naming conventions:
  - `<domain>.event.<action>`
  - `<domain>.cmd.<action>`
- Include dead-letter and retry strategy in shared messaging utilities.

## gRPC Plan

- Use gRPC for synchronous service-to-service calls.
- Define gRPC contracts in root `proto/`.
- Generate and consume stubs from `shared/proto/`.
- Keep handlers in each service, not in `shared/`.

## Local Development and Ops Plan

- Use `Tiltfile` for local compile/build/deploy loops.
- Keep dev/prod manifests under `infra/development/k8s` and `infra/production/k8s`.
- Keep Dockerfiles under corresponding `infra/*/docker`.
- Include infra resources for RabbitMQ and tracing from day one.

## Testing Plan

- Unit tests colocated with implementation (`*_test.go`).
- Prefer table-driven tests for business and validation logic.
- Add integration tests for PostgreSQL repositories, RabbitMQ publishers/consumers, and gRPC boundaries where critical.
- Keep fixtures explicit and minimal.

## Implementation Sequence

1. Bootstrap the monorepo skeleton.
2. Add root `proto/` and codegen into `shared/proto/`.
3. Implement `users-service` with PostgreSQL + `goose` + `sqlc`.
4. Implement `catalog-service` with PostgreSQL + `goose` + `sqlc`.
5. Add RabbitMQ shared messaging and the first event-driven flow.
6. Implement `inventory-service`, then `orders-service`.
7. Add `payment-service` and complete the order-payment async workflow.
8. Harden observability, retries, dead-lettering, and integration tests.

## Long-Term Goal

Each service should be independently buildable and deployable, while the repository keeps a consistent structure, shared contracts, and reproducible local/dev/prod workflows.
