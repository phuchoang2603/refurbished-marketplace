# External Secrets

## Purpose

Define how External Secrets Operator and Doppler sync application credentials into Talos GitOps clusters without committing plaintext secrets.

## Requirements

### Requirement: Doppler SecretStore with service token

Application ExternalSecrets SHALL reference the marketplace-owned namespaced SecretStore named doppler in ecommerce. It SHALL authenticate using ecommerce/doppler-token, key dopplerToken. This repository SHALL render the SecretStore but SHALL NOT contain a real token value. Marketplace ExternalSecrets SHALL NOT reference the platform `ClusterSecretStore/doppler`, which serves talos-proxmox secrets such as the Cloudflare Tunnel token.

#### Scenario: Store references bootstrap secret

- **WHEN** a fresh workload cluster is prepared
- **THEN** the ecommerce SecretStore is Ready using the environment-appropriate bootstrap token before application ExternalSecret sync

#### Scenario: Service token not in Git

- **WHEN** the repository is cloned
- **THEN** no Doppler token value is present in tracked files

#### Scenario: Marketplace does not use the platform store

- **WHEN** marketplace charts are rendered
- **THEN** every ExternalSecret references `SecretStore/doppler` in its own namespace and none references a `ClusterSecretStore`

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

### Requirement: No committed plaintext cluster secrets

The repository SHALL NOT commit plaintext Kubernetes Secret manifests for application credentials. `infra/k8s/secrets.yaml` SHALL be removed.

#### Scenario: Bootstrap without secrets.yaml

- **WHEN** External Secrets Operator syncs after Doppler bootstrap
- **THEN** application secrets are created by ESO and not from a committed `infra/k8s/secrets.yaml`

### Requirement: Doppler bootstrap secret manifests

The repository SHALL provide example Kubernetes Secret manifests for Doppler service tokens and document manual creation of gitignored `dev` and `prd` bootstrap files under `infra/k8s/`.

#### Scenario: Developer creates talos-dev bootstrap secret

- **WHEN** a developer prepares talos-dev
- **THEN** they copy `infra/k8s/doppler-token.dev.secret.yaml.example` to `infra/k8s/doppler-token.dev.secret.yaml` and paste a `dev` config service token

#### Scenario: Operator bootstraps prod cluster secret

- **WHEN** the prod cluster is prepared for GitOps
- **THEN** an operator applies `infra/k8s/doppler-token.prd.secret.yaml` manually before application `ExternalSecret` resources sync

### Requirement: devenv Doppler CLI defaults

The repository SHALL provide Doppler CLI via devenv and set `DOPPLER_PROJECT` and `DOPPLER_CONFIG` in `devenv.nix` for local secret management.

#### Scenario: devenv configures Doppler CLI context

- **WHEN** a developer enters `devenv shell`
- **THEN** `DOPPLER_PROJECT` and `DOPPLER_CONFIG` are available for `doppler` CLI commands

#### Scenario: Bootstrap applies Doppler secret

- **WHEN** `infra/k8s/doppler-token.dev.secret.yaml` is applied on talos-dev
- **THEN** Kubernetes Secret `doppler-token` exists in `ecommerce` with key `dopplerToken`

### Requirement: Provider swap via SecretStore

Secret provisioning SHALL remain provider-agnostic at the service deployment layer. Changing the external secrets provider SHALL require updating the marketplace-owned SecretStore and, if remote key names change, marketplace chart `externalSecrets` / service `db` / `auth` settings — not service deployment templates.

#### Scenario: Deployment templates unchanged after provider swap

- **WHEN** the `SecretStore` provider is changed from Doppler to another supported ESO provider
- **THEN** `refurbished-marketplace` service deployments and the `kafka` chart continue referencing the same Kubernetes Secret names

### Requirement: Doppler environment configs

Doppler MAY use separate configs for non-production vs production secrets. Bootstrap of `ecommerce/doppler-token` SHALL use kubectl (or equivalent) with the target environment's kubeconfig, fetched from the `talos-proxmox` Doppler project config for that environment.

#### Scenario: Bootstrap with kubectl

- **WHEN** a contributor bootstraps secrets for dev or prod
- **THEN** they fetch that environment's `KUBECONFIG` from Doppler project `talos-proxmox` and create `ecommerce/doppler-token` with kubectl against that cluster

### Requirement: Mongo credentials from Doppler

The repository SHALL render ExternalSecret resources that populate Kubernetes Secrets for MongoDB Community SCRAM users (and any replica-set key material the operator requires) from Doppler via the `ecommerce/doppler` SecretStore. Plaintext Mongo passwords SHALL NOT be committed.

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

### Requirement: Products Mongo credentials from Doppler

The repository SHALL sync Doppler-backed credentials that products and the products Mongo outbox connector use to authenticate to the catalog replica set (SCRAM user password or equivalent). Plaintext Mongo URIs with passwords SHALL NOT be committed. Products SHALL NOT keep a Postgres `products-app` database secret after `products_db` is removed.

#### Scenario: Products can authenticate to Mongo

- **WHEN** ExternalSecrets have synced successfully after cutover
- **THEN** products and the Kafka Connect service account can read a Secret in `ecommerce` with the keys needed to connect to the catalog replica set

#### Scenario: Products Postgres secret is gone

- **WHEN** the marketplace chart no longer deploys `products_db`
- **THEN** GitOps SHALL NOT render a products CloudNativePG app username/password ExternalSecret

### Requirement: Inventory database secret from Doppler

The marketplace chart SHALL render an ExternalSecret for inventory’s CNPG credentials from Doppler using the same `{SECRET_NAME}_PASSWORD` pattern as other service databases. Plaintext inventory passwords SHALL NOT be committed.

#### Scenario: Inventory app secret exists

- **WHEN** ExternalSecrets have synced successfully
- **THEN** an inventory app Secret exists in `ecommerce` with keys usable by CloudNativePG and the inventory Deployment

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

### Requirement: Single SecretStore owner

`SecretStore/doppler` in `ecommerce` SHALL be rendered by exactly one chart, deployed as its own `secret-store` child Application with a sync wave earlier than every Application that renders ExternalSecrets. No other marketplace chart SHALL render a SecretStore.

#### Scenario: One Application tracks the store

- **WHEN** the dev or prod root has synced
- **THEN** `SecretStore/doppler` is tracked only by the `<env>-secret-store` Application and no `SharedResourceWarning` is reported for it

#### Scenario: Store failure is isolated

- **WHEN** `ecommerce/doppler-token` is missing or invalid
- **THEN** the `secret-store` Application reports the unhealthy store, and consumer charts do not render or own the store themselves
