## MODIFIED Requirements

### Requirement: Release workflow publishes Docker images to GHCR

The repository SHALL build and publish the complete current image matrix for image, chart, Argo root or delivery workflow changes. Documentation-only changes SHALL NOT trigger delivery.

#### Scenario: Push to main with image-related changes

- **WHEN** a commit is pushed to `main` and modifies image or deployment configuration paths
- **THEN** the release workflow runs and builds and pushes all configured images in the release matrix

#### Scenario: Push to main without image-related changes

- **WHEN** a commit is pushed to `main` without modifying image or deployment configuration paths
- **THEN** the release workflow is not triggered

#### Scenario: Manual release of all images

- **WHEN** the release workflow is triggered via `workflow_dispatch`
- **THEN** it builds and pushes all configured images in the release matrix

### Requirement: GHCR image naming and tags

Published images SHALL use PR tags `pr-<number>-<merge-sha>` and main tags `main-<sha>`, coordinated across the complete matrix. Image repositories SHALL use the naming pattern `ghcr.io/<repository>/<image>` where `<repository>` is `${{ github.repository }}` and `<image>` is the short service image name (for example `web`, `users-migrator`, `connect-debezium`).

#### Scenario: Image tagged on publish

- **WHEN** the release workflow publishes an image for commit `abc123`
- **THEN** the image is available at `ghcr.io/<repository>/<image>:main-abc123` for a main release

### Requirement: Pull request image publish

Trusted repository PRs SHALL publish immutable PR image tags for dev deployment. Fork PRs SHALL build without registry or deployment credentials. PRs SHALL NOT overwrite main release tags.

#### Scenario: PR push publishes SHA tags

- **WHEN** a trusted repository pull request changes image or delivery configuration paths
- **THEN** CI publishes the PR merge revision tag for every image in the build matrix

#### Scenario: PR does not overwrite main

- **WHEN** the pull request image job runs
- **THEN** it does not push the `:main` tag

#### Scenario: Dispatch on a non-main ref is not prod

- **WHEN** the image job runs via `workflow_dispatch` on a ref other than `refs/heads/main`
- **THEN** it does not publish or deploy production images

### Requirement: Delete PR-only GHCR versions after the PR closes

Closing a PR SHALL NOT delete its GHCR versions while dev can still reference them. No automatic close-triggered cleanup SHALL remove the active dev release.

#### Scenario: Closed PR cleanup

- **WHEN** a pull request closes while dev still references its published images
- **THEN** those images remain available for pod restart and replacement

### Requirement: Release workflow includes all infra docker images

The release workflow SHALL build and push all marketplace and infra images declared in the `release-images.yml` matrix on every workflow run that executes the release job, including all eight application services and `connect-debezium`. Cluster deploys SHALL pull those GHCR tags via Argo CD.

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
