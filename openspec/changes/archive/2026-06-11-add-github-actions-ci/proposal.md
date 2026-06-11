## Why

Issue #4 calls for automated CI, but the repository has no GitHub Actions workflows today. Local quality gates (devenv, treefmt, codegen scripts) are not enforced on push/PR, and integration tests that rely on Testcontainers are only run manually. A focused CI workflow closes the first slice of #4 without blocking on image publishing or ArgoCD.

## What Changes

- Add `.github/workflows/ci.yml` that runs on pull requests and pushes to `main`.
- Add a **lint** job across all Go modules using `golangci-lint`, `go vet`, and `go build`.
- Add **path-filtered integration test jobs** per service module, with **shared dependency fan-out** so changes under `shared/` retest affected services.
- Include **`services/web` tests** (httptest-based; no Testcontainers) when web paths or shared dependencies change.
- Add a **Helm validation** job (`helm lint`, `helm template`, `kubeconform`) that runs only when `infra/charts/**` changes.
- Add repository **golangci-lint configuration** aligned with the multi-module layout.

### Non-goals (this change)

- Container image build/push to GHCR (deferred to a follow-up PR).
- treefmt / codegen drift checks in CI (remain local devenv/git-hook responsibilities).
- ArgoCD bootstrap, deploy runbooks, or migration promotion strategy.
- CI for raw manifests under `infra/k8s/` (e.g. payment-gateway-simulator).

## Capabilities

### New Capabilities

- `github-actions-ci`: Automated GitHub Actions validation for Go lint, selective service integration tests, and Helm chart rendering/conformance when chart files change.

### Modified Capabilities

- _(none)_

## Impact

- **New files**: `.github/workflows/ci.yml`, `.golangci.yml` (or equivalent config path used by the workflow).
- **CI runtime**: GitHub-hosted runners with Docker available for Testcontainers-backed service tests.
- **Contributor workflow**: PRs receive lint and targeted test status checks; chart edits trigger Helm validation.
- **Issue #4**: Partially satisfies acceptance criteria for workflows on PR/main and selective `go test`; image push and ArgoCD items remain out of scope.
