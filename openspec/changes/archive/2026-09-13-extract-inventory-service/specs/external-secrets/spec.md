## ADDED Requirements

### Requirement: Inventory database secret from Doppler

The marketplace chart SHALL render an ExternalSecret for inventory’s CNPG credentials from Doppler using the same `{SECRET_NAME}_PASSWORD` pattern as other service databases. Plaintext inventory passwords SHALL NOT be committed.

#### Scenario: Inventory app secret exists

- **WHEN** ExternalSecrets have synced successfully
- **THEN** an inventory app Secret exists in `ecommerce` with keys usable by CloudNativePG and the inventory Deployment
