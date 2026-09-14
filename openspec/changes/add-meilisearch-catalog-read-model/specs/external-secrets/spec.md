## ADDED Requirements

### Requirement: Meilisearch master key from Doppler

The repository SHALL sync a Doppler-backed master key that Meilisearch and the search service use. Plaintext Meilisearch keys SHALL NOT be committed.

#### Scenario: Search secret exists in ecommerce

- **WHEN** ExternalSecrets for Meilisearch have synced successfully
- **THEN** a Secret in `ecommerce` exists with the key Meilisearch and the search service need to authenticate to the search HTTP API

#### Scenario: No Meilisearch keys in Git

- **WHEN** the repository is cloned
- **THEN** no Meilisearch master key values are present in tracked files

### Requirement: Secrets documentation lists Meilisearch Doppler keys

Secrets documentation SHALL list the Doppler remote key and Kubernetes Secret name used for Meilisearch.

#### Scenario: Operator can bootstrap the Meilisearch key

- **WHEN** an operator prepares Doppler configs `dev` and `prd`
- **THEN** documentation names the Meilisearch Doppler key to set before the search Application can become Ready
