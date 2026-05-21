## REMOVED Requirements

### Requirement: Products is catalog only

**Reason**: Catalog, stock, and reservation persistence are being reunited under one service boundary so listing reads and writes no longer need a separate inventory service.
**Migration**: Move product, inventory, and reservation ownership into the unified catalog boundary and stop treating stock as a separate downstream concern.

## MODIFIED Requirements

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for stock-aware product reads and unified listing creation within the catalog boundary.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching product or not-found

## ADDED Requirements

### Requirement: Products owns colocated stock state

The products service MUST own product catalog data together with the colocated stock state needed for marketplace listing reads and writes.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches product data for a detail or admin-oriented stock-aware catalog flow
- **THEN** the service SHALL return product data from the unified catalog boundary without requiring a separate inventory service lookup

#### Scenario: Product list is read

- **WHEN** a caller fetches a catalog product list in the first merged phase
- **THEN** the service SHALL allow that list flow to stay lighter than detail/admin surfaces and SHALL NOT require exact stock quantities everywhere

### Requirement: Products creates listings with initial stock in one logical operation

The products service MUST support creating a product and initializing its stock state within one logical catalog write path, and it MUST require initial stock explicitly in that create operation.

#### Scenario: Listing is created

- **WHEN** a caller creates a new product listing with an initial stock quantity
- **THEN** the service SHALL persist the product record and its initial stock state without requiring a second downstream service call

#### Scenario: Listing is created without initial stock

- **WHEN** a caller attempts to create a new product listing without an explicit initial stock quantity
- **THEN** the service SHALL reject the request instead of silently defaulting stock
