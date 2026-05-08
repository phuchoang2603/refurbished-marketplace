## ADDED Requirements

### Requirement: Cart state is ephemeral

The cart service MUST store only session/cart state and MUST NOT persist cart data in PostgreSQL.

#### Scenario: Cart is loaded

- **WHEN** a client loads a cart
- **THEN** the service SHALL read cart state from Redis or Valkey

#### Scenario: Cart is cleared

- **WHEN** an order is created successfully
- **THEN** the service SHALL clear the cart state for that cart

### Requirement: Cart state expires automatically

The cart service MUST apply TTL-based expiration to abandoned carts.

#### Scenario: Cart is abandoned

- **WHEN** a cart is left unused past its TTL
- **THEN** the stored cart state SHALL expire automatically
