## ADDED Requirements

### Requirement: Product records carry category and condition

The products service MUST persist a `category` from a fixed set (`phone`, `laptop`, `tablet`, `accessory`, `headphones`, `smartwatch`, `game_console`, `camera`) and a `condition` grade (`excellent`, `good`, `fair`) on product creation, reject other values, and return both on every read including batch lookup.

#### Scenario: Listing with an unknown category

- **WHEN** a caller creates a product with a category outside the fixed set
- **THEN** the service rejects it without persisting the listing

#### Scenario: Batch read returns grading

- **WHEN** checkout batch-reads products
- **THEN** each returned product carries its category and condition
