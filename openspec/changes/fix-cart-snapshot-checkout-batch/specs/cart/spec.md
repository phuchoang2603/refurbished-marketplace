## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Cart supports multi-item remove

The cart service MUST support removing multiple product lines in a single operation so checkout can clear a merchant group without N sequential Redis document rewrites.

#### Scenario: Several product IDs are removed together

- **WHEN** a caller invokes multi-remove with a valid cart identifier and a non-empty set of product identifiers
- **THEN** the service SHALL load the cart once, remove every item whose product_id is in that set, persist once, and return the updated cart

#### Scenario: Multi-remove includes product IDs not in the cart

- **WHEN** a multi-remove request includes product identifiers that are not present on the cart
- **THEN** the service SHALL ignore those identifiers and still succeed for the remaining removals (idempotent multi-remove)

#### Scenario: Single-item remove remains available

- **WHEN** a caller removes one product_id via the existing single-item remove API
- **THEN** the service SHALL remove that line and return the updated cart as today
