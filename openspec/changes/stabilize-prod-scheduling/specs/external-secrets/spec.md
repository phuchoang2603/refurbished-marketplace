## ADDED Requirements

### Requirement: Single SecretStore owner

`SecretStore/doppler` in `ecommerce` SHALL be rendered by exactly one chart, deployed as its own `secret-store` child Application with a sync wave earlier than every Application that renders ExternalSecrets. No other marketplace chart SHALL render a SecretStore.

#### Scenario: One Application tracks the store

- **WHEN** the dev or prod root has synced
- **THEN** `SecretStore/doppler` is tracked only by the `<env>-secret-store` Application and no `SharedResourceWarning` is reported for it

#### Scenario: Store failure is isolated

- **WHEN** `ecommerce/doppler-token` is missing or invalid
- **THEN** the `secret-store` Application reports the unhealthy store, and consumer charts do not render or own the store themselves
