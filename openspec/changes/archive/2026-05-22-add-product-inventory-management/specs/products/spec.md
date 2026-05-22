## MODIFIED Requirements

### Requirement: Products creates listings with initial stock in one logical operation

The products service MUST support authenticated seller-managed listing creation through one logical catalog write path that persists the product record together with explicit initial stock.

#### Scenario: Seller-managed listing is created

- **WHEN** a trusted internal caller creates a product for an authenticated seller-managed listing
- **THEN** the service SHALL persist the catalog fields and initial stock for that product in one logical operation

#### Scenario: Seller-managed listing is created without explicit stock

- **WHEN** a caller attempts to create a seller-managed listing without explicit initial stock
- **THEN** the service SHALL reject the request instead of silently defaulting stock

## ADDED Requirements

### Requirement: Product records carry seller ownership

The products service MUST persist the seller ownership identifier provided as `merchant_id` on product creation so downstream order and payment flows can attribute the listing consistently.

#### Scenario: Product is created with a merchant owner

- **WHEN** a caller creates a product with a valid `merchant_id`
- **THEN** the service SHALL store that `merchant_id` with the catalog record and return it in subsequent reads
