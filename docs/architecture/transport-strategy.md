# Transport Strategy

## Decision

Adopt a strict transport boundary:

- REST only at the edge (`web` service)
- gRPC for internal service-to-service communication

This applies to `users`, `products`, and `orders`.

## Why

- Keeps external API ergonomics simple for browser/mobile clients.
- Keeps internal contracts strongly typed and versioned via protobuf.
- Makes service boundaries explicit and easier to evolve.

## Service Roles

### web (edge)

- Owns all client-facing REST routes.
- Handles request/response shaping for frontend use cases.
- Calls internal services using gRPC clients.
- Enforces REST authorization (JWT validation + route protection).

### users/products/orders (internal)

- Expose gRPC services and protobuf contracts.
- Own domain logic, DB access, migrations, and events.
- Do not add new public REST APIs.

## Migration Path from Current State

Migration is now complete for users transport boundary. Current order going forward:

1. Web REST endpoints call users gRPC for auth and user lookup.
2. Users REST handlers are removed; users is internal gRPC only.
3. Apply same pattern for products and orders.

## API Ownership

- REST schema ownership: `web` service
- gRPC schema ownership: each internal service under `services/<service>/proto/v1/`

## Related Architecture

- Synchronous internal calls: gRPC
- Asynchronous integration: RabbitMQ events
- Mesh policy (Istio): authn/authz/traffic controls, but not token issuance or refresh business logic

## Authorization Boundary

- Authentication session lifecycle remains in `users` service.
- Authorization for REST route access is enforced in `web`.
- Domain authorization invariants (e.g. product ownership for mutations) are enforced in the owning service as well.
