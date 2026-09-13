## ADDED Requirements

### Requirement: Inventory participates in service CI

CI SHALL include inventory in the lint module list, service path filters and outputs, test matrix, and vulnerability-scan matrix using the same triggers and shared dependency rules as other Go services.

#### Scenario: Inventory changes are checked

- **WHEN** inventory code or its shared dependencies change in a pull request
- **THEN** inventory lint, tests, and vulnerability scanning SHALL run

#### Scenario: Scheduled vulnerability scan

- **WHEN** the scheduled full vulnerability scan runs
- **THEN** inventory SHALL be scanned alongside the other services
