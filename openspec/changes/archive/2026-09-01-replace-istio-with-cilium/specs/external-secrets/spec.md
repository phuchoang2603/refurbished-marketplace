## MODIFIED Requirements

### Requirement: External Secrets Operator installed

The repository SHALL install External Secrets Operator on the Talos cluster using the upstream Helm chart in the `operators` namespace via Argo CD. Tilt SHALL NOT install ESO.

#### Scenario: ESO operator healthy after GitOps sync

- **WHEN** the Talos operators Application syncs
- **THEN** the External Secrets Operator deployment becomes ready in the `operators` namespace

### Requirement: Doppler environment configs

Doppler MAY use separate configs for non-production vs production secrets. Bootstrap of `operators/doppler-token` SHALL use kubectl (or equivalent) as on staging. A Tilt-applied `doppler-token.dev.secret.yaml` and Colima-only bootstrap SHALL NOT be required.

#### Scenario: Bootstrap without Tilt

- **WHEN** a contributor bootstraps secrets after this change
- **THEN** they create `operators/doppler-token` without running `tilt up`
