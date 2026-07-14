# GHCR Release

## Purpose

Define how container images under `infra/docker/` are built and published to GitHub Container Registry so staging and production can pin coordinated image tags.

## Requirements

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

### Requirement: GHCR image naming and tags

Published images SHALL use the naming pattern `ghcr.io/<repository>/<image>` where `<repository>` is `${{ github.repository }}` and `<image>` is the short service image name (for example `web`, `users-migrator`, `connect-debezium`).

#### Scenario: Image tagged on publish

- **WHEN** the release workflow publishes an image for commit `abc123`
- **THEN** the image is available at `ghcr.io/<repository>/<image>:abc123` and `ghcr.io/<repository>/<image>:main`

### Requirement: Release workflow includes all infra docker images

The release workflow SHALL build and push all marketplace and infra images declared in the `release-images.yml` matrix on every workflow run that executes the release job, including application services, migrators, `payment-gateway-simulator`, and `connect-debezium`. Local Colima image builds for marketplace services SHALL be driven by Tilt `docker_build` rules.

#### Scenario: Full image matrix on main push

- **WHEN** the release workflow runs for a push to `main` that triggers the workflow
- **THEN** it builds and pushes every image listed in the workflow matrix `include`

#### Scenario: Full image matrix on manual dispatch

- **WHEN** the release workflow runs via `workflow_dispatch`
- **THEN** it builds and pushes every image listed in the workflow matrix `include`

### Requirement: Release workflow uses standard GitHub Actions

The release workflow SHALL use `docker/login-action` and `docker/build-push-action` (or equivalent official Docker GitHub Actions) with `GITHUB_TOKEN` for GHCR authentication.

#### Scenario: GHCR login before push

- **WHEN** the release workflow pushes an image
- **THEN** it authenticates to `ghcr.io` using the workflow `GITHUB_TOKEN` with `packages: write` permission
