# Continuous integration

GitHub Actions runs `.github/workflows/ci.yml` on every pull request and push to `main`.

| Job             | When it runs      | What it does                                                                                         |
| --------------- | ----------------- | ---------------------------------------------------------------------------------------------------- |
| `lint`          | Always            | `devenv tasks run repo:lint`: ast-grep rules, `scripts/lint.sh` (golangci-lint, helm lint, conftest) |
| `test` (matrix) | Path filter match | `scripts/check.sh <service>` for each affected service                                               |
| `govulncheck`   | Weekly schedule   | `scripts/vulncheck.sh` across all Go modules                                                         |

Every job runs inside devenv through the shared reusable workflow, so CI uses the same toolchain as local development. Import boundaries (`depguard` in `devenv.nix`) and chart policies (`policy/`) are enforced by `lint`.

**Branch protection:** require the `lint` job. Service test jobs are not created when a PR does not touch relevant paths — that is expected.

Treefmt's existing Oxfmt integration formats plain YAML locally. Helm templates remain excluded. The marketplace repository does not validate or deploy the platform-owned Cloudflare Tunnel chart.

## Container images (GHCR)

CI validates charts and builds only images whose service dependencies changed since their last deployed image commit. Shared code and build-input changes rebuild their consumers; chart-only changes reuse all image tags. The first release builds every image. Trusted PRs publish `pr-<number>-<merge-sha>`; main releases publish `main-<sha>`. Fork PRs build without publishing or deployment credentials.

After all affected images publish, CI updates `deploy/dev` for a trusted PR or `deploy/prod` for main. The deployment commit pins child Applications to the built source SHA and per-service commit tags. Unchanged services retain their existing tags in `infra/argocd/app-of-apps/release.yaml`; no digest manifest is maintained. Each locally bootstrapped Argo root watches its own deployment ref, so CI needs only registry/Git access and never Kubernetes or Doppler credentials. Publishing skips source refs that moved and uses force-with-lease to protect concurrent updates. Deployment refs are generated release snapshots and should not be edited manually.

Closing a PR does not delete its images: dev may still need them for pod restart until the next successful release. No workflow writes image-pin commits to main.
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

Formatting and protobuf codegen (`treefmt`, `codegen:proto`) stay out of CI; `repo:lint` fails when sqlc output is stale. Changes to `devenv.*` trigger every service.
