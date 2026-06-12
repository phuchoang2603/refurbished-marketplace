## ADDED Requirements

### Requirement: External Secrets Operator installed

The repository SHALL install External Secrets Operator on the local Tilt Kubernetes cluster using the upstream Helm chart in the `external-secrets` namespace.

#### Scenario: ESO operator healthy on tilt up

- **WHEN** a developer runs `tilt up` with a valid cluster context
- **THEN** the External Secrets Operator deployment becomes ready in the `external-secrets` namespace

### Requirement: Doppler ClusterSecretStore with service token

The repository SHALL configure a `ClusterSecretStore` that authenticates to Doppler using a service token stored in Kubernetes Secret `doppler-token` with key `dopplerToken` in the `external-secrets` namespace.

#### Scenario: Store references bootstrap secret

- **WHEN** ESO evaluates the Doppler `ClusterSecretStore`
- **THEN** it reads the service token from `external-secrets/doppler-token` key `dopplerToken`

#### Scenario: Service token not in Git

- **WHEN** the repository is cloned
- **THEN** no Doppler service token value is present in tracked files

### Requirement: ExternalSecrets sync chart secrets

The repository SHALL define `ExternalSecret` resources that sync Doppler secrets into Kubernetes Secrets in the `ecommerce` namespace with these names and keys: `users-app`, `products-app`, `orders-app`, `payment-app` (username/password), and `users-auth` (JWT_SECRET).

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

### Requirement: devenv and Tilt Doppler bootstrap

The repository SHALL provide Doppler CLI via devenv, load `DOPPLER_TOKEN` from gitignored `.env` through devenv dotenv, and bootstrap the ESO auth secret from Tilt using that environment variable.

#### Scenario: devenv loads service token

- **WHEN** a developer enters `devenv shell` with `DOPPLER_TOKEN` in `.env`
- **THEN** `DOPPLER_TOKEN` is available to Tilt child processes

#### Scenario: Tilt creates doppler-token secret

- **WHEN** `tilt up` runs with `DOPPLER_TOKEN` set
- **THEN** Tilt applies or updates Kubernetes Secret `doppler-token` in `external-secrets` with key `dopplerToken`

### Requirement: Provider swap via ClusterSecretStore

Secret provisioning SHALL remain provider-agnostic at the Helm and application layers. Changing the external secrets provider SHALL require updating ESO store configuration only, not service Helm templates.

#### Scenario: Helm unchanged after provider swap

- **WHEN** the `ClusterSecretStore` provider is changed from Doppler to another supported ESO provider
- **THEN** `refurbished-marketplace` and `kafka` charts continue referencing the same Kubernetes Secret names

### Requirement: Doppler environment configs

Doppler SHALL use separate configs for local development and production (for example `dev` and `prd`). Local Tilt SHALL use a Doppler service token scoped to the development config.

#### Scenario: Local dev config

- **WHEN** bootstrapping local development
- **THEN** the service token used for ESO is scoped to the Doppler development config only
