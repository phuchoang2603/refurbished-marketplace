# Transport Strategy

## Decision

Adopt a strict transport boundary:

- REST only at the edge (`web` service)
- gRPC for internal service-to-service communication

This applies to `users`, `products`, `inventory`, and `orders`.

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

### users/products/inventory/orders (internal)

- Expose gRPC services and protobuf contracts.
- Own domain logic, DB access, migrations, and events.
- Do not add new public REST APIs.

### products scope

- Public reads only in the web API.
- Product catalog writes stay internal/admin-only.
- Inventory is a separate internal service and should not be folded into products.

## Migration Path from Current State

Migration is now complete for users transport boundary. Current order going forward:

1. Web REST endpoints call users gRPC for auth and user lookup.
2. Users REST handlers are removed; users is internal gRPC only.
3. Apply same pattern for products, inventory, and orders.

## API Ownership

- REST schema ownership: `web` service
- gRPC schema ownership: each internal service under `services/<service>/proto/v1/`

### gRPC Contracts and Clients (Current)

- Users protobuf contract is centralized at `shared/proto/users/v1/users.proto`.
- Generated users gRPC code lives in `shared/proto/users/v1/`.
- Reusable users gRPC client lives in `shared/proto/usersclient/`.
- Products protobuf contract is centralized at `shared/proto/products/v1/products.proto`.
- Generated products gRPC code lives in `shared/proto/products/v1/`.
- Reusable products gRPC client lives in `shared/proto/productsclient/`.
- Orders protobuf contract is centralized at `shared/proto/orders/v1/orders.proto`.
- Generated orders gRPC code lives in `shared/proto/orders/v1/`.
- Reusable orders gRPC client lives in `shared/proto/ordersclient/`.
- Inventory protobuf contract lives at `shared/proto/inventory/v1/inventory.proto`.
- Generated inventory gRPC code lives in `shared/proto/inventory/v1/`.

## Related Architecture

- Synchronous internal calls: gRPC
- Asynchronous integration: Kafka events (Strimzi preferred for Kubernetes)
- Mesh policy (Istio): authn/authz/traffic controls, but not token issuance or refresh business logic

## Authorization Boundary

- Authentication session lifecycle remains in `users` service.
- Authorization for REST route access is enforced in `web`.
- Domain authorization invariants (e.g. product ownership for mutations) are enforced in the owning service as well.

## Product and Order Scope

- Scope is normal ecommerce, not C2C escrow.
- Recommender is a later consumer of event history.
- External payment/fraud platform is a separate project and integrates through service APIs/events.
