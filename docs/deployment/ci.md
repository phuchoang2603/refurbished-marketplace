# Continuous integration

GitHub Actions runs `.github/workflows/ci.yml` on every pull request and push to `main`.

| Job             | When it runs                           | What it does                                                       |
| --------------- | -------------------------------------- | ------------------------------------------------------------------ |
| `lint`          | Always                                 | `golangci-lint` via Reviewdog (blocking PR comments)               |
| `govulncheck`   | Path filter match (or weekly schedule) | `govulncheck` SARIF uploaded to Code Scanning per affected service |
| `test` (matrix) | Path filter match                      | `go test ./...` for the affected service module                    |
| `helm`          | `infra/charts/**` changed              | `helm lint`, `helm template`, and `kubeconform`                    |

On pull requests, Reviewdog posts inline review comments for `golangci-lint` findings on changed lines.

The `govulncheck` job uses the same service matrix and path-filter fan-out as `test`. SARIF results are uploaded to GitHub Code Scanning and do not fail CI. A weekly schedule runs a full scan across all services.

**Branch protection:** require the `lint` job. Service test jobs and `helm` may be skipped when a PR does not touch relevant paths — that is expected.

## Container images (GHCR)

Pushes to `main` that touch image-related paths trigger `.github/workflows/release-images.yml` (`:main` and `:<commit-sha>`). Pull requests on the same paths push `:<commit-sha>` and `:pr-<n>` (never `:main`). Closed PRs run `.github/workflows/cleanup-pr-images.yml` to delete `:pr-<n>` package versions (not `:main`). Argo CD `global.imageTag` must be the git SHA, not the moving `pr-*` tag. Retarget the cluster to `main` before cleanup.

See [gitops.md](gitops.md).

## Path-filter fan-out for tests

| Changed paths                 | Tests triggered                             |
| ----------------------------- | ------------------------------------------- |
| `services/<name>/**`          | That service only                           |
| `shared/proto/**`             | users, products, orders, cart, payment, web |
| `shared/auth/**`              | users, web                                  |
| `shared/messaging/**`         | products, orders, payment                   |
| `shared/err/dberr/**`         | users, products, orders, payment            |
| `shared/err/grpcerr/**`       | users, products, orders, cart, payment      |
| `shared/runtime/**`           | users, products, orders, cart, payment, web |
| `shared/observe/log/**`       | users, products, orders, cart, payment, web |
| `shared/observe/trace/**`     | products, orders, payment, web              |
| `shared/testutil/postgres/**` | users, products, orders, payment            |
| `shared/testutil/kafka/**`    | products, orders, payment                   |
| `shared/testutil/redis/**`    | cart                                        |

Local formatting and codegen drift checks (`treefmt`, `generate-proto`, `sqlc-gen`) stay out of CI.
