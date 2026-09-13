## MODIFIED Requirements

### Requirement: Products owns colocated stock state

The products service MUST own listing identity and catalog fields. It MUST NOT persist available or reserved quantity. It MUST NOT call the inventory service. Product reads SHALL return catalog fields only.

#### Scenario: Product is read with stock summary

- **WHEN** a caller fetches a product by id for detail or cart stamping
- **THEN** the service SHALL return catalog fields from the catalog store and SHALL NOT load stock from inventory

#### Scenario: Product list is read

- **WHEN** a caller fetches a catalog product list before Meilisearch exists
- **THEN** the list MAY be empty; the service SHALL NOT require SQL `products` OFFSET listing after the catalog table is removed

### Requirement: Products creates listings with initial stock in one logical operation

The products service MUST persist the catalog record for seller-managed listing creation. It MUST NOT seed inventory. Explicit initial stock is required at the web/inventory boundary, not inside products.

#### Scenario: Seller-managed listing is created

- **WHEN** a trusted internal caller creates a product for an authenticated seller-managed listing
- **THEN** the service SHALL persist the catalog fields and return the listing identity without writing stock

#### Scenario: Seller-managed listing is created without explicit stock

- **WHEN** a caller attempts to create a seller-managed listing without explicit initial stock at the web boundary
- **THEN** web SHALL reject the request instead of silently defaulting stock; products SHALL NOT invent a quantity

### Requirement: Products exposes internal gRPC methods

The products service MUST expose internal gRPC methods for catalog reads and listing creation and deletion. It MUST NOT expose ReserveStock.

#### Scenario: Product lookup occurs

- **WHEN** a caller requests a product by ID
- **THEN** the service SHALL return the matching catalog product or not-found

#### Scenario: Reserve is requested over gRPC

- **WHEN** a documented internal caller requests reservation for an order on the products API
- **THEN** products SHALL NOT apply the reserve; reservation SHALL be served by inventory

#### Scenario: Reserve is requested on products

- **WHEN** a caller invokes reservation on the products API
- **THEN** the method is absent; reservation SHALL go to inventory

### Requirement: Products supports batch lookup by IDs

The products service MUST expose a batch read that returns catalog rows for a set of product identifiers so checkout can re-validate many lines without N sequential single-get RPCs. The batch SHALL NOT join inventory.

#### Scenario: Multiple known IDs are requested

- **WHEN** a caller requests products by a non-empty list of product IDs within the service’s allowed batch size
- **THEN** the service SHALL return catalog records for every ID that exists, without stock summaries from the inventory ledger

#### Scenario: Some IDs are missing

- **WHEN** a batch request includes product IDs that do not exist in the catalog
- **THEN** the service SHALL omit those IDs from the response rather than failing the entire batch solely because some IDs are missing

#### Scenario: Batch size exceeds the limit

- **WHEN** a caller requests more product IDs than the configured maximum batch size
- **THEN** the service SHALL reject the request as invalid

## ADDED Requirements

### Requirement: Products deletes a listing

The products service MUST delete a catalog listing by id so web can compensate a failed EnsureStock.

#### Scenario: Compensating delete after inventory seed fails

- **WHEN** web requests delete for a listing id it just created because EnsureStock failed
- **THEN** the listing SHALL be removed from the catalog store and SHALL NOT remain visible on subsequent Get by id

## REMOVED Requirements

### Requirement: Products manages reservations

**Reason**: Reservations belong to the inventory runtime.
**Migration**: Call inventory ReserveStock and consume inventory Kafka from the inventory service.

### Requirement: Products consumes order item events

**Reason**: Reservation Kafka consumers run in inventory.
**Migration**: Point the consumer group and handlers at `services/inventory`.

### Requirement: Products reserves stock on an internal command

**Reason**: ReserveStock moves to inventory.v1.
**Migration**: Web and any other callers use inventory gRPC with the same idempotent full-order semantics.
