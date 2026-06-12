## Why

Issue #4 requires publishing container images from `infra/docker/` to GHCR after CI landed for lint, tests, and Helm validation. Images are built locally via Tilt today but are not available to non-local clusters. A dedicated release workflow closes the next acceptance criterion while keeping local Tilt unchanged.

## What Changes

- Add `.github/workflows/release-images.yml` to build and push all 12 images under `infra/docker/` to GHCR on path-filtered pushes to `main`.
- Tag images as `ghcr.io/${{ github.repository }}/<image>:<sha>` (plus optional rolling `main` tag).
- Update `.github/workflows/ci.yml` lint job to post PR review comments via official Reviewdog actions for `golangci-lint` and `govulncheck`.
- Make `govulncheck` informational only (never fails the lint job).
- Revert Go toolchain from **1.26.4** back to **1.26.2** across `go.work`, modules, CI, and Go Dockerfiles.

### Non-goals

- ArgoCD bootstrap, env-specific Helm values, or cluster deploy wiring.
- Helm chart changes to consume GHCR image refs (follow-up for issue #4).
- Custom JavaScript PR comment scripts.
- Forcing CI failure on vulnerability findings from `govulncheck`.

## Capabilities

### New Capabilities

- `ghcr-release`: Path-filtered GitHub Actions workflow that builds and pushes all `infra/docker/*.Dockerfile` images to GHCR with versioned tags.

### Modified Capabilities

- `github-actions-ci`: Lint job uses Reviewdog PR comments; `govulncheck` is non-blocking; Go toolchain reverted to 1.26.2. Container publishing remains in the separate `ghcr-release` workflow (not in `ci.yml`).

## Impact

- **New**: `.github/workflows/release-images.yml`
- **Updated**: `.github/workflows/ci.yml`, `go.work`, all `go.mod` files, Go service Dockerfiles, `CONTRIBUTING.md`
- **Issue #4**: Partially satisfies “CI builds and pushes versioned images to a container registry”
- **Tilt/local dev**: Unchanged local image names; `GOWORK=off` in Dockerfiles preserved
