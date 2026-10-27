# Continuous integration

GitHub Actions runs `.github/workflows/ci.yml` on every pull request and push to `main`.

| Job             | When it runs                           | What it does                                                       |
| --------------- | -------------------------------------- | ------------------------------------------------------------------ |
| `lint`          | Always                                 | `golangci-lint` via Reviewdog (blocking PR comments)               |
| `govulncheck`   | Path filter match (or weekly schedule) | `govulncheck` SARIF uploaded to Code Scanning per affected service |
| `test` (matrix) | Path filter match                      | `go test ./...` for the affected service module                    |

On pull requests, Reviewdog posts inline review comments for `golangci-lint` findings on changed lines.

The `govulncheck` job uses the same service matrix and path-filter fan-out as `test`. SARIF results are uploaded to GitHub Code Scanning and do not fail CI. A weekly schedule runs a full scan across all services.

**Branch protection:** require the `lint` job. Service test jobs may be skipped when a PR does not touch relevant paths — that is expected.

Treefmt's existing Oxfmt integration formats plain YAML locally. Helm templates remain excluded. The marketplace repository does not validate or deploy the platform-owned Cloudflare Tunnel chart.

## Container images (GHCR)

Each event builds every service image once. Pull requests (same-repo only; fork PRs build without publishing) push `:pr-<number>-<full-merge-sha>`, a new push cancels the previous build, and closing the PR cancels any active build. Pushes to `main` that touch image-related paths push `:main-<full-sha>` and `:latest`, then the `pin` job commits `global.imageTag: main-<full-sha>` to the production values of the `refurbished-marketplace` and `kafka` charts (`[skip ci]`). `.github/workflows/cleanup-pr-images.yml` runs when a PR closes and deletes only versions tagged exclusively with `pr-<number>-*`; `main-*` and `latest` are never selected. Dev charts default to `:latest`.

See [gitops.md](gitops.md).

## Path-filter fan-out for tests

| Changed paths                    | Tests triggered                                       |
| -------------------------------- | ----------------------------------------------------- |
| `services/<name>/**`             | That service only                                     |
| `shared/proto/**`                | users, products, search, inventory, orders, cart, web |
| `shared/auth/**`                 | users, web                                            |
| `shared/messaging/**`            | products, inventory, orders, search                   |
| `shared/err/dberr/**`            | users, inventory, orders                              |
| `shared/err/grpcerr/**`          | users, products, inventory, orders, cart, search      |
| `shared/runtime/**`              | users, products, inventory, orders, cart, web, search |
| `shared/observe/log/**`          | users, products, inventory, orders, cart, web, search |
| `shared/observe/trace/**`        | products, inventory, orders, web, search              |
| `shared/testutil/postgres/**`    | users, products, inventory, orders                    |
| `shared/testutil/kafka/**`       | products, inventory, orders, search, checkout         |
| `shared/testutil/meilisearch/**` | search                                                |
| `shared/testutil/redis/**`       | cart, checkout                                        |

Local formatting and codegen drift checks (`treefmt`, `codegen:proto`, `codegen:sqlc`) stay out of CI.
