## ADDED Requirements

### Requirement: Mongo credentials from Doppler

The repository SHALL render ExternalSecret resources that populate Kubernetes Secrets for MongoDB Community SCRAM users (and any replica-set key material the operator requires) from Doppler via the existing `doppler` ClusterSecretStore. Plaintext Mongo passwords SHALL NOT be committed.

#### Scenario: SCRAM secret exists in ecommerce

- **WHEN** ExternalSecrets for Mongo have synced successfully
- **THEN** a Secret in `ecommerce` exists with the keys the MongoDB Community custom resource `passwordSecretRef` expects

#### Scenario: No Mongo passwords in Git

- **WHEN** the repository is cloned
- **THEN** no Mongo root or app password values are present in tracked files

### Requirement: Secrets documentation lists Mongo Doppler keys

Secrets documentation SHALL list the Doppler remote keys and Kubernetes Secret names used for Mongo.

#### Scenario: Operator can bootstrap Mongo passwords

- **WHEN** an operator prepares Doppler configs `dev` and `prd`
- **THEN** documentation names the Mongo-related Doppler keys to set before the database Application can become Ready
