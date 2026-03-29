# Refurbished Marketplace Plan

## Guiding Shape

This project should stay close to the Boot.dev style: small packages, explicit SQL, simple handlers, and table-driven tests. The repo should also follow the broad structure of `microservices-go-starter`: `services/`, `shared/`, `docs/`, and `infra/` as the maintop-level areas.

## Canonical Repo Tree

```text
/
  PLAN.md
  flake.nix
  go.mod
  go.sum
  services/
    catalog/
      cmd/catalog/
      internal/
      db/
        migrations/
        queries/
      proto/
    users/
    orders/
    inventory/
    payments/
    events/
  shared/
    contracts/
    env/
    retry/
    util/
  docs/
    architecture/
  infra/
    development/
    production/
  tools/
```

## Service Layout Rules

- Each service owns its own business logic and persistence.
- Put the executable in `cmd/<service>/`.
- Put request handling, domain logic, and repositories under `internal/`.
- Keep gRPC definitions in `proto/` when a service exposes or consumes gRPC APIs.
- Keep migrations and SQL queries next to the service that owns the schema.

## Suggested Service Split

- `catalog`: product listings, refurb condition, pricing, and search-facing data
- `users`: accounts, profiles, auth-related user state
- `inventory`: stock levels, item availability, reservation lifecycle
- `orders`: checkout, order state, and coordination across services
- `payments`: payment intent and transaction records
- `events`: event publishing/consuming helpers and contract definitions

## Database Plan

- Use PostgreSQL as the source of truth for each service.
- Use `goose` for schema migrations.
- Use `sqlc` for typed query generation.
- Favor plain SQL files over ORM layers.
- Keep one schema owner per service to avoid shared-database coupling.

## Messaging Plan

- Use RabbitMQ for asynchronous workflows and cross-service events.
- Treat events as integration boundaries, not as a replacement for service APIs.
- Standardize event naming early, such as `catalog.item.created` or `orders.order.completed`.
- Keep message payloads small and versionable.

## gRPC Plan

- Use gRPC for synchronous service-to-service calls where latency and contract clarity matter.
- Keep protobuf files minimal and versioned by service.
- Generate code into service-local output directories or a shared generated area if needed.

## Testing Plan

- Put unit tests next to implementation files.
- Use table-driven tests for parsing, validation, and business rules.
- Add integration tests for PostgreSQL and RabbitMQ interactions.
- Keep test fixtures tiny and explicit.
- Prefer testing behavior through public service methods and handlers.

## Development Plan

2. Create the canonical repo tree and shared top-level folders.
3. Scaffold the first service with `cmd/`, `internal/`, `db/`, and `proto/`.
4. Add PostgreSQL migrations and `sqlc` queries for the first service.
5. Add RabbitMQ publishing/consuming for one workflow.
6. Add gRPC contracts for one synchronous service boundary.
7. Build out tests as each module lands.

## Implementation Order

- Start with `users` and `catalog` so the marketplace has identity and listings.
- Add `inventory` next so availability can be tracked explicitly.
- Add `orders` after the core data model is stable.
- Add `payments` and event-driven flows once order state exists.

## Code Style Notes

- Keep files short and focused.
- Prefer explicit dependencies over hidden globals.
- Use standard library types and helpers first.
- Keep error handling direct and readable.
- Avoid over-abstracting until duplication is real.

## Long-Term Goal

Reach a point where each service can be built, tested, and deployed independently while still following one consistent project shape.
