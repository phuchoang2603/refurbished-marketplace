## ADDED Requirements

### Requirement: Products supports batch lookup by IDs

The products service MUST expose a batch read that returns authoritative catalog rows for a set of product identifiers so marketplace composition (especially checkout) can re-validate many lines without N sequential single-get RPCs.

#### Scenario: Multiple known IDs are requested

- **WHEN** a caller requests products by a non-empty list of product IDs within the service’s allowed batch size
- **THEN** the service SHALL return product records for every ID that exists, from the PostgreSQL write model (including stock summary fields consistent with single-get)

#### Scenario: Some IDs are missing

- **WHEN** a batch request includes product IDs that do not exist
- **THEN** the service SHALL omit those IDs from the response rather than failing the entire batch solely because some IDs are missing

#### Scenario: Batch size exceeds the limit

- **WHEN** a caller requests more product IDs than the configured maximum batch size
- **THEN** the service SHALL reject the request as invalid
