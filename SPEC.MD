# Refurbished Marketplace Spec

## Project Goal

Build a microservice refurbished marketplace backend in Go. The system should support marketplace workflows such as listing inventory, managing users, handling orders, and coordinating cross-service events through asynchronous messaging.

## Core Stack

- Language: Go
- Go dependencies: standard library + `google.golang.org/grpc`
- Database: PostgreSQL
- Event bus: RabbitMQ

## Code Style

Prefer the style used in the Boot.dev Go exercises in this repo:

- Keep packages small and focused
- Use clear, simple names over clever abstractions
- Prefer the standard library first
- Use table-driven tests for behavior-heavy code
- Keep functions short and direct
- Organize code so the happy path is easy to read

## Repository Layout

Try to mirror the `microservices-go-starter` structure as much as practical:

- `services/` for deployable microservices.
- `shared/` for common helpers, contracts, and reusable code
- `docs/` for architecture notes and flow diagrams
- `infra/` for local and production deployment assets
- `sql/` or service-local migrations/queries for database work

## Folder Scaffolding

Use a consistent service layout so each microservice feels familiar:

- `services/<service-name>/cmd/<service-name>/` for the executable entrypoint
- `services/<service-name>/internal/` for private service logic
- `services/<service-name>/internal/handlers/` for HTTP or gRPC handlers
- `services/<service-name>/internal/service/` for business logic
- `services/<service-name>/internal/repository/` for persistence code
- `services/<service-name>/proto/` for gRPC definitions when needed
- `services/<service-name>/db/migrations/` for schema migrations
- `services/<service-name>/db/queries/` for SQL query files if using generated DB access
- `services/<service-name>/tests/` only for higher-level integration tests when colocated tests are not enough

Keep shared code limited to reusable contracts, env helpers, retry logic, and common utilities.

## Database Workflow

- `goose` for versioned PostgreSQL migrations
- `sqlc` for generating typed Go query code from SQL
- Keep SQL in files instead of hiding it behind a large ORM
- Use generated code for database access, but keep the surrounding service code hand-written and simple

## Testing Layout

- Put unit tests next to the code they verify
- Prefer `*_test.go` files in the same package unless there is a strong reason not to
- Use table-driven tests for validation, parsing, and business logic
- Add integration tests for database and message-bus behavior where useful
- Keep test fixtures minimal and explicit

## Service Design Notes

- Use gRPC for service-to-service communication where it makes sense
- Use PostgreSQL for durable business data
- Use RabbitMQ for domain events and asynchronous workflows
- Keep service boundaries explicit so each service owns its data and behavior

## Initial Expectations

- Favor simple, readable implementations over framework-heavy ones
- Introduce shared code only when multiple services truly need it
- Document major flows as the system grows
