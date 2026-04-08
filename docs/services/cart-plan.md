# Cart Service Plan

## Goal

Keep cart state ephemeral and local to the cart service so checkout can remain a clean boundary.

## Scope

- Service: `services/cart`
- Transport: gRPC
- Storage: Redis/Valkey

## Behavior

- Cart state should be keyed by a stable `cart_id`.
- Guest carts and logged-in carts should follow the same flow.
- Clear the cart after successful order creation.
- Use TTL so abandoned carts expire automatically.

## Data

- Store only session/cart state.
- Keep product price snapshots for display only.
- Do not persist cart state in Postgres.

## Runtime

- Cart should connect to Redis via `REDIS_ADDR`.
- Tests should use Redis testcontainers or a compatible substitute.
