## MODIFIED Requirements

### Requirement: Release workflow includes all infra docker images

The release workflow SHALL build and push all marketplace and infra images declared in the `release-images.yml` matrix on every workflow run that executes the release job, including application services (including `search`), migrators, `payment-gateway-simulator`, and `connect-debezium`. Cluster deploys SHALL pull those GHCR tags via Argo CD.

#### Scenario: Full image matrix on main push

- **WHEN** the release workflow runs for a push to `main` that triggers the workflow
- **THEN** it builds and pushes every image listed in the workflow matrix `include`

#### Scenario: Full image matrix on manual dispatch

- **WHEN** the release workflow runs via `workflow_dispatch`
- **THEN** it builds and pushes every image listed in the workflow matrix `include`
