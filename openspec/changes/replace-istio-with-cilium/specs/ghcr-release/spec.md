## ADDED Requirements

### Requirement: Pull request image publish

The repository SHALL build and push marketplace/infra images to GHCR on pull requests so Argo CD can deploy a branch by git SHA before merge. PR workflows SHALL NOT retag `:main`.

#### Scenario: PR push publishes SHA tags

- **WHEN** a pull request changes image-related paths (same path set as the main release workflow, or a documented subset)
- **THEN** CI pushes `ghcr.io/<repository>/<image>:<git-sha>` for the PR head commit for each image in the build matrix

#### Scenario: PR does not overwrite main

- **WHEN** the pull request image job runs
- **THEN** it does not push the `:main` tag

#### Scenario: Dispatch on a non-main ref is not prod

- **WHEN** the image job runs via `workflow_dispatch` on a ref other than `refs/heads/main`
- **THEN** it tags `:<git-sha>` only, not `:main`

### Requirement: Delete PR-only GHCR versions after the PR closes

After a pull request is merged or closed, the repository SHALL delete GHCR package versions tagged with that PR’s commit SHAs. It SHALL NOT delete `:main` or versions that also carry the `main` tag.

#### Scenario: Closed PR cleanup

- **WHEN** a pull request is closed (merged or not)
- **THEN** a workflow deletes GHCR versions whose tags include a SHA from that PR and do not include `main`

## MODIFIED Requirements

### Requirement: Release workflow includes all infra docker images

The release workflow SHALL build and push all marketplace and infra images declared in the `release-images.yml` matrix on every workflow run that executes the release job, including application services, migrators, `payment-gateway-simulator`, and `connect-debezium`. Cluster deploys SHALL pull those GHCR tags via Argo CD. Tilt `docker_build` SHALL NOT be required for Talos.

#### Scenario: Full image matrix on main push

- **WHEN** the release workflow runs for a push to `main` that triggers the workflow
- **THEN** it builds and pushes every image listed in the workflow matrix `include`

#### Scenario: Full image matrix on manual dispatch

- **WHEN** the release workflow runs via `workflow_dispatch`
- **THEN** it builds and pushes every image listed in the workflow matrix `include`
