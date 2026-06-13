## MODIFIED Requirements

### Requirement: Release workflow publishes Docker images to GHCR

The repository SHALL provide a GitHub Actions workflow that builds and pushes container images from Dockerfiles under `infra/docker/` to GitHub Container Registry.

#### Scenario: Push to main with image-related changes

- **WHEN** a commit is pushed to `main` and modifies paths under `infra/docker/**`, `services/**`, `shared/**`, `tools/**`, `go.work`, or any `go.mod`
- **THEN** the release workflow runs and builds and pushes all configured images in the release matrix

#### Scenario: Push to main without image-related changes

- **WHEN** a commit is pushed to `main` without modifying image-related paths
- **THEN** the release workflow is not triggered

#### Scenario: Manual release of all images

- **WHEN** the release workflow is triggered via `workflow_dispatch`
- **THEN** it builds and pushes all configured images in the release matrix

## REMOVED Requirements

### Requirement: Per-image path-filter fan-out

**Reason:** GitOps production deploys pin a single commit SHA via `global.imageTag`; every service image must receive the same `:sha` tag on each main merge.

**Migration:** Remove path-filter outputs and conditional matrix skips from `release-images.yml`; CI test path filters remain unchanged in `ci.yml`.

## MODIFIED Requirements

### Requirement: Release workflow includes all infra docker images

The release workflow SHALL build and push all twelve Dockerfiles under `infra/docker/` on every workflow run that executes the release matrix, including application services, migrators, `payment-gateway-simulator`, and `connect-debezium`.

#### Scenario: Full image matrix on main push

- **WHEN** the release workflow runs for a push to `main` that triggers the workflow
- **THEN** it builds and pushes all twelve images defined under `infra/docker/*.Dockerfile`

#### Scenario: Full image matrix on manual dispatch

- **WHEN** the release workflow runs via `workflow_dispatch`
- **THEN** it builds and pushes all twelve images defined under `infra/docker/*.Dockerfile`
