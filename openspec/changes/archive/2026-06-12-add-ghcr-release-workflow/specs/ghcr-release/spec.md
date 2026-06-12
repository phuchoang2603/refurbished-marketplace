## ADDED Requirements

### Requirement: Release workflow publishes Docker images to GHCR

The repository SHALL provide a GitHub Actions workflow that builds and pushes container images from every Dockerfile under `infra/docker/` to GitHub Container Registry.

#### Scenario: Push to main with image-related changes

- **WHEN** a commit is pushed to `main` and modifies paths under `infra/docker/**`, `services/**`, `shared/**`, `tools/**`, `go.work`, or any `go.mod`
- **THEN** the release workflow builds and pushes all configured service images to GHCR

#### Scenario: Push to main without image-related changes

- **WHEN** a commit is pushed to `main` without modifying image-related paths
- **THEN** the release workflow is not triggered

### Requirement: GHCR image naming and tags

Published images SHALL use the naming pattern `ghcr.io/<repository>/<image>` where `<repository>` is `${{ github.repository }}` and `<image>` is the short service image name (for example `web`, `users-migrator`, `connect-debezium`).

#### Scenario: Image tagged on publish

- **WHEN** the release workflow publishes an image for commit `abc123`
- **THEN** the image is available at `ghcr.io/<repository>/<image>:abc123`

### Requirement: Release workflow includes all infra docker images

The release workflow SHALL build and push images for all twelve Dockerfiles under `infra/docker/`, including application services, migrators, `payment-gateway-simulator`, and `connect-debezium`.

#### Scenario: Full image matrix on release

- **WHEN** the release workflow runs
- **THEN** it builds and pushes all twelve images defined under `infra/docker/*.Dockerfile`

### Requirement: Release workflow uses standard GitHub Actions

The release workflow SHALL use `docker/login-action` and `docker/build-push-action` (or equivalent official Docker GitHub Actions) with `GITHUB_TOKEN` for GHCR authentication.

#### Scenario: GHCR login before push

- **WHEN** the release workflow pushes an image
- **THEN** it authenticates to `ghcr.io` using the workflow `GITHUB_TOKEN` with `packages: write` permission
