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
  services/
    users/
    products/
    orders/
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
```

## Service Layout Rules

- Start simple: services may begin with a flat layout when small.
- Move to `cmd/`, `internal/`, and optional `pkg/` only when complexity grows.
- Keep service business logic private under `internal/`.
- Keep transport handlers close to service code (HTTP, gRPC, event consumers).
- Each service owns its persistence and migrations.

## Proto and Codegen Plan

- Keep protobuf definitions service-local first: `services/<service>/proto/v1/`.
- Promote to a root-level `proto/` only when contracts are shared across multiple services.
- Generate Go code into service-local output first; introduce `shared/proto/` only when reuse is proven.
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
- Define contracts in each service under `services/<service>/proto/v1/`.
- Generate and consume stubs from the owning service until shared contracts are needed.
- Keep handlers in each service, not in `shared/`.

## Local Development and Ops Plan

- Use `Tiltfile` for local compile/build/deploy loops.
- Keep dev/prod manifests under `infra/development/k8s` and `infra/production/k8s`.
- Keep Dockerfiles under corresponding `infra/*/docker`.
- Include infra resources for RabbitMQ and tracing from day one.
- Plan to use Istio ingress and traffic policies for edge routing first.

## Testing Plan

- Unit tests colocated with implementation (`*_test.go`).
- Prefer table-driven tests for business and validation logic.
- Add integration tests for PostgreSQL repositories, RabbitMQ publishers/consumers, and gRPC boundaries where critical.
- Keep fixtures explicit and minimal.

## Implementation Sequence

1. Bootstrap the monorepo skeleton.
2. Add service-local `proto/v1` folders and per-service codegen.
3. Implement `users` with PostgreSQL + `goose` + `sqlc`.
4. Implement `products` with PostgreSQL + `goose` + `sqlc`.
5. Implement `orders` and connect the first event-driven workflow.
6. Add RabbitMQ shared messaging where async coordination is needed.
7. Add optional future services (for example payments) in separate projects or later phases.
8. Harden observability, retries, dead-lettering, and integration tests.

## Long-Term Goal

Each service should be independently buildable and deployable, while the repository keeps a consistent structure, shared contracts, and reproducible local/dev/prod workflows.
