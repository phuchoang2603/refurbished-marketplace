## MODIFIED Requirements

### Requirement: Doppler ClusterSecretStore with service token

Application ExternalSecrets SHALL reference the platform-owned ClusterSecretStore named doppler. The platform prerequisite SHALL authenticate using operators/doppler-token, key dopplerToken. This repository SHALL NOT render the store or contain a real token value.

#### Scenario: Store references bootstrap secret

- **WHEN** a fresh workload cluster is prepared
- **THEN** the platform store is Ready using the environment-appropriate bootstrap token before application ExternalSecrets sync

#### Scenario: Service token not in Git

- **WHEN** the repository is cloned
- **THEN** no Doppler token value is present in tracked files

### Requirement: Provider swap via ClusterSecretStore

Secret provisioning SHALL remain provider-agnostic at the service deployment layer. Changing the external secrets provider SHALL require updating the platform-owned store configuration in talos-proxmox and, if remote key names change, marketplace chart `externalSecrets` / service `db` / `auth` settings — not service deployment templates.

#### Scenario: Deployment templates unchanged after provider swap

- **WHEN** the `ClusterSecretStore` provider is changed from Doppler to another supported ESO provider
- **THEN** `refurbished-marketplace` service deployments and the `kafka` chart continue referencing the same Kubernetes Secret names

### Requirement: ExternalSecrets sync chart secrets

The repository SHALL render `ExternalSecret` resources from the `refurbished-marketplace` Helm chart for each enabled service with `db` (basic-auth username/password) and for each unique `auth.secretName` (for example `users-auth` / `JWT_SECRET`). Doppler remote password keys SHALL follow `{SECRET_NAME}_PASSWORD` derived from `db.secretName`. The Kubernetes Secret username SHALL be generated from `db.owner`, defaulting to `<service>_app`; no Doppler username key is required.

#### Scenario: DB secret available for CNPG

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `users-app` exists in `ecommerce` with `username` and `password` keys usable by CloudNativePG and service Helm templates

#### Scenario: JWT secret available for web and users

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `users-auth` exists in `ecommerce` with `JWT_SECRET` key

#### Scenario: Debezium connector secrets

- **WHEN** ExternalSecrets have synced successfully
- **THEN** orders, payment, and inventory Postgres credentials and catalog Mongo credentials exist under the names referenced by Strimzi connectors; no obsolete products Postgres credential is required

## REMOVED Requirements

### Requirement: External Secrets Operator installed

**Reason**: ESO installation belongs to talos-proxmox; this repository retains application ExternalSecrets only.

**Migration**: None. Fresh installations consume the platform provided by talos-proxmox; existing resource adoption is outside scope.
