## ADDED Requirements

### Requirement: Products is catalog only

The products service MUST manage catalog data and MUST NOT own inventory or reservation state.

#### Scenario: Product is created

- **WHEN** an internal/admin caller creates a product
- **THEN** the service SHALL persist catalog fields for that product

#### Scenario: Product is read

- **WHEN** a client fetches catalog data
- **THEN** the service SHALL return product details without inventory mutation

### Requirement: Products exposes internal gRPC methods

The products service MUST expose gRPC methods for product creation and reads.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching product or not-found
