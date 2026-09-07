# cart Specification

## Purpose

Define ephemeral merchant-aware cart state, expiration, and removal of purchased product lines.

## Requirements

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

#### Scenario: Cart is cleared

- **WHEN** payment succeeds for an order and the caller requests removal of its product lines
- **THEN** the service SHALL remove those paid lines while preserving other cart items

### Requirement: Cart state expires automatically

The cart service MUST apply TTL-based expiration to abandoned carts.

#### Scenario: Cart is abandoned

- **WHEN** a cart is left unused past its TTL
- **THEN** the stored cart state SHALL expire automatically

### Requirement: Cart items carry merchant-aware checkout state

The cart capability MUST store and return `merchant_id` alongside `product_id` and `quantity` for each cart item so checkout can group items by merchant without reconstructing merchant boundaries from scratch. Each cart item MUST also carry a display snapshot of product presentation fields used by the cart UI: at least `product_name` and `unit_price_cents` supplied by the caller on write (replication into ephemeral cart state, not a catalog SoR).

#### Scenario: Item is added to cart

- **WHEN** a caller adds an item to the cart with a product identifier, merchant identifier, quantity, product name, and unit price in cents
- **THEN** the cart state SHALL persist the merchant-aware item shape including those display snapshot fields in ephemeral storage

#### Scenario: Existing line is incremented

- **WHEN** a caller adds quantity for a product_id already present in the cart with a new display snapshot
- **THEN** the cart SHALL increase quantity and SHALL replace the stored product name and unit price with the values from that request

#### Scenario: Merchant identifier is omitted from item write

- **WHEN** a caller attempts to add or update a cart item without a merchant identifier
- **THEN** the cart service SHALL reject the request as invalid

#### Scenario: Display snapshot is omitted or invalid on item write

- **WHEN** a caller attempts to add or update a cart item without a non-empty product name or with a non-positive unit price in cents
- **THEN** the cart service SHALL reject the request as invalid

#### Scenario: Cart is read

- **WHEN** a caller loads an existing cart
- **THEN** the returned cart SHALL include the stored `merchant_id`, `product_name`, and `unit_price_cents` for each cart item (zero or empty snapshot fields only when historical state predated snapshots)

### Requirement: Cart supports multi-item remove

The cart service MUST support removing multiple product lines in a single operation so checkout can clear a merchant group without N sequential Redis document rewrites.

#### Scenario: Cart supports removing one or many product IDs

- **WHEN** a caller invokes remove with a valid cart identifier and a non-empty set of product identifiers (one or more)
- **THEN** the service SHALL load the cart once, remove every item whose product_id is in that set, persist once, and return the updated cart

#### Scenario: Multi-remove includes product IDs not in the cart

- **WHEN** a remove request includes product identifiers that are not present on the cart
- **THEN** the service SHALL ignore those identifiers and still succeed for the remaining removals (idempotent remove)
