## MODIFIED Requirements

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

### Requirement: devenv Doppler CLI defaults

The repository SHALL provide Doppler CLI via devenv and set `DOPPLER_PROJECT` and `DOPPLER_CONFIG` in `devenv.nix` for local secret management.

#### Scenario: devenv configures Doppler CLI context

- **WHEN** a developer enters `devenv shell`
- **THEN** `DOPPLER_PROJECT` and `DOPPLER_CONFIG` are available for `doppler` CLI commands

#### Scenario: Bootstrap applies Doppler secret

- **WHEN** `infra/k8s/doppler-token.dev.secret.yaml` is applied on talos-dev
- **THEN** Kubernetes Secret `doppler-token` exists in `ecommerce` with key `dopplerToken`

### Requirement: Doppler environment configs

Doppler MAY use separate configs for non-production vs production secrets. Bootstrap of `ecommerce/doppler-token` SHALL use kubectl (or equivalent) with the target environment's kubeconfig, fetched from the `talos-proxmox` Doppler project config for that environment.

#### Scenario: Bootstrap with kubectl

- **WHEN** a contributor bootstraps secrets for dev or prod
- **THEN** they fetch that environment's `KUBECONFIG` from Doppler project `talos-proxmox` and create `ecommerce/doppler-token` with kubectl against that cluster
