## ADDED Requirements

### Requirement: Cart items carry merchant-aware checkout state

The cart capability MUST store and return `merchant_id` alongside `product_id` and `quantity` for each cart item so checkout can group items by merchant without reconstructing merchant boundaries from scratch.

#### Scenario: Item is added to cart

- **WHEN** a caller adds an item to the cart with a product identifier, merchant identifier, and quantity
- **THEN** the cart state SHALL persist the merchant-aware item shape in ephemeral storage

#### Scenario: Merchant identifier is omitted from item write

- **WHEN** a caller attempts to add or update a cart item without a merchant identifier
- **THEN** the cart service SHALL reject the request as invalid

#### Scenario: Cart is read

- **WHEN** a caller loads an existing cart
- **THEN** the returned cart SHALL include the stored `merchant_id` for each cart item
