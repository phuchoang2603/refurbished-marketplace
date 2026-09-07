## MODIFIED Requirements

### Requirement: Cart state is ephemeral

The cart service MUST store only session/cart state and MUST NOT persist cart data in PostgreSQL.

#### Scenario: Cart is loaded

- **WHEN** a client loads a cart
- **THEN** the service SHALL read cart state from Redis or Valkey

#### Scenario: Cart lines are removed after payment succeeds

- **WHEN** payment for a merchant-scoped order succeeds and the caller requests multi-remove of that order's product IDs
- **THEN** the service SHALL remove those product IDs from the cart document when they are still present

#### Scenario: Order is placed but unpaid

- **WHEN** an order is placed successfully and hosted payment has not succeeded
- **THEN** the service SHALL NOT require those product IDs to already be absent from the cart
