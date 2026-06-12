## 1. Go toolchain revert

- [x] 1.1 Revert `go.work`, all `go.mod` files, and Go service Dockerfiles from Go 1.26.4 to Go 1.26.2
- [x] 1.2 Run `tidy` (or `go work sync`) and verify modules resolve at 1.26.2

## 2. CI lint — Reviewdog comments

- [x] 2.1 Update `ci.yml` lint job permissions for Reviewdog (`pull-requests: write`, `checks: write`)
- [x] 2.2 Replace manual golangci-lint step with `reviewdog/action-golangci-lint@v2` (`fail_level: error`, `filter_mode: diff_context`, PR `github-pr-review`)
- [x] 2.3 Add informational `govulncheck` step via `reviewdog/action-setup@v1` + Reviewdog CLI with `-fail-level=none` and `continue-on-error: true`
- [x] 2.4 Set `GO_VERSION` to 1.26.2 in `ci.yml`

## 3. GHCR release workflow

- [x] 3.1 Create `.github/workflows/release-images.yml` triggered on path-filtered push to `main` and `workflow_dispatch`
- [x] 3.2 Add matrix for all 12 `infra/docker/*.Dockerfile` images with context `.`
- [x] 3.3 Use `docker/login-action@v3` and `docker/build-push-action@v6` to push `ghcr.io/${{ github.repository }}/<image>:${{ github.sha }}` and `:main`
- [x] 3.4 Grant `packages: write` only on the release workflow

## 4. Documentation and verification

- [x] 4.1 Update `CONTRIBUTING.md` with GHCR release workflow and lint comment behavior
- [x] 4.2 Verify release workflow YAML and matrix image list match Tilt/Dockerfile inventory
