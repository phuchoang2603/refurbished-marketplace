## ADDED Requirements

### Requirement: External Secrets Operator installed

The repository SHALL install External Secrets Operator on the local Tilt Kubernetes cluster using the upstream Helm chart in the `operators` namespace.

#### Scenario: ESO operator healthy on tilt up

- **WHEN** a developer runs `tilt up` with a valid cluster context
- **THEN** the External Secrets Operator deployment becomes ready in the `operators` namespace

### Requirement: Doppler ClusterSecretStore with service token

The repository SHALL configure a `ClusterSecretStore` that authenticates to Doppler using a service token stored in Kubernetes Secret `doppler-token` with key `dopplerToken` in the `operators` namespace.

#### Scenario: Store references bootstrap secret

- **WHEN** ESO evaluates the Doppler `ClusterSecretStore`
- **THEN** it reads the service token from `operators/doppler-token` key `dopplerToken`

#### Scenario: Service token not in Git

- **WHEN** the repository is cloned
- **THEN** no Doppler service token value is present in tracked files

### Requirement: ExternalSecrets sync chart secrets

The repository SHALL render `ExternalSecret` resources from the `refurbished-marketplace` Helm chart for each enabled service with `db` (basic-auth username/password) and for each unique `auth.secretName` (for example `users-auth` / `JWT_SECRET`). Doppler remote keys for DB secrets SHALL follow `{SECRET_NAME}_USERNAME` and `{SECRET_NAME}_PASSWORD` derived from `db.secretName`.

#### Scenario: DB secret available for CNPG

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `users-app` exists in `ecommerce` with `username` and `password` keys usable by CloudNativePG and service Helm templates

#### Scenario: JWT secret available for web and users

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `users-auth` exists in `ecommerce` with `JWT_SECRET` key

#### Scenario: Debezium connector secrets

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `orders-app` and `payment-app` secrets exist for Strimzi `${secrets:…}` references in the kafka chart

### Requirement: No committed plaintext cluster secrets

The repository SHALL NOT commit plaintext Kubernetes Secret manifests for application credentials. `infra/k8s/secrets.yaml` SHALL be removed.

#### Scenario: Tilt without secrets.yaml

- **WHEN** a developer runs `tilt up` after bootstrap
- **THEN** application secrets are created by ESO and not from `k8s_yaml('./infra/k8s/secrets.yaml')`

### Requirement: devenv Doppler bootstrap

The repository SHALL provide Doppler CLI via devenv, set `DOPPLER_PROJECT` and `DOPPLER_CONFIG` in `devenv.nix`, load `DOPPLER_TOKEN` from gitignored `.env` through devenv dotenv, and generate `infra/k8s/doppler-token.secret.yaml` via devenv `files` when the token is set.

#### Scenario: devenv loads service token

- **WHEN** a developer enters `devenv shell` with `DOPPLER_TOKEN` in `.env`
- **THEN** `DOPPLER_TOKEN` is available to Tilt child processes

#### Scenario: devenv links doppler-token manifest

- **WHEN** `devenv shell` runs with `DOPPLER_TOKEN` set
- **THEN** `infra/k8s/doppler-token.secret.yaml` is linked with key `dopplerToken` for the Doppler ClusterSecretStore

#### Scenario: Tilt applies cluster bootstrap manifests

- **WHEN** `tilt up` runs with the doppler-token manifest present
- **THEN** Kubernetes Secret `doppler-token` exists in `operators` with key `dopplerToken`

### Requirement: Provider swap via ClusterSecretStore

Secret provisioning SHALL remain provider-agnostic at the service deployment layer. Changing the external secrets provider SHALL require updating `infra/k8s/cluster-secret-store.yaml` and, if remote key names change, marketplace chart `externalSecrets` / service `db` / `auth` settings — not service deployment templates.

#### Scenario: Deployment templates unchanged after provider swap

- **WHEN** the `ClusterSecretStore` provider is changed from Doppler to another supported ESO provider
- **THEN** `refurbished-marketplace` service deployments and the `kafka` chart continue referencing the same Kubernetes Secret names

### Requirement: Doppler environment configs

Doppler SHALL use separate configs for local development and production (for example `dev` and `prd`). Local Tilt SHALL use a Doppler service token scoped to the development config.

#### Scenario: Local dev config

- **WHEN** bootstrapping local development
- **THEN** the service token used for ESO is scoped to the Doppler development config only
