# Contributing

Thanks for helping build this project. This guide covers local development, how work is planned, and how GitHub issues and PRs fit together.

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A local Kubernetes runtime used by Tilt (for example Colima + Docker)
- [Tilt](https://tilt.dev/) (provided inside the devenv shell)

## Local development

### Enter the development shell

This repository uses `devenv` to install and pin local tooling. Enter the development shell before running generators, tests, or local infrastructure commands:

```bash
devenv shell
```

The shell provides the project tooling defined in `devenv.nix`, including Go, protobuf tooling, database migration/query generators, Kubernetes tooling, Tilt, and OpenSpec.

### Run the stack with Tilt

Local Kubernetes development is managed with Tilt. After entering the `devenv` shell, start the stack with:

```bash
tilt up
```

Tilt uses the root `Tiltfile` to build services, apply the Kubernetes/Helm resources under `infra/`, and keep the local cluster in sync while you edit code. Use the Tilt UI to inspect service status, logs, resource readiness, and rebuilds.

Integration tests rely on Testcontainers for Kafka, PostgreSQL, and Redis/Valkey. Prefer verifying behavior through Tilt for full-service flows; run targeted Go tests when they add meaningful coverage.

### Code generation

From inside `devenv shell`:

| Command          | Purpose                                                                  |
| ---------------- | ------------------------------------------------------------------------ |
| `generate-proto` | Regenerate Go code from `**/proto/*/v1/*.proto`                          |
| `sqlc-gen`       | Regenerate sqlc query code for services with `sqlc.yaml`                 |
| `templ generate` | Regenerate `templ` views (run from `services/web` when templates change) |
| `tidy`           | `go mod tidy` across modules and `go work sync` for `go.work`            |

Edit SQL migrations under `services/<service>/db/migrations/` and queries under `services/<service>/db/queries/`, then run `sqlc-gen` when query shapes change.

Formatting is handled by devenv/git hooks via `treefmt` (`gofumpt`, `sqruff`, etc.).

Local Go development uses a root `go.work` so tools like gopls and `golangci-lint` see the whole repo. Container builds keep `ENV GOWORK=off` and copy only the service plus required `shared/` paths, so images still build a single module.

## How work is planned (OpenSpec)

OpenSpec is the authoritative planning workflow for non-trivial changes. Active work lives under `openspec/changes/<change-name>/` with artifacts such as:

- `proposal.md` — why and what changes
- `design.md` — decisions and trade-offs
- `specs/` — delta requirements by capability
- `tasks.md` — implementation checklist

Typical flow:

1. **Propose** a change (new directory under `openspec/changes/`).
2. **Implement** against `tasks.md`, updating specs/design when reality diverges.
3. **Verify** implementation against artifacts before archive.
4. **Archive** with spec sync: `openspec archive <change-name> -y`

Archived changes move to `openspec/changes/archive/YYYY-MM-DD-<change-name>/`. Main specs live in `openspec/specs/`.

Cursor command shortcuts for this workflow live under `.cursor/commands/opsx-*.md` if you use Cursor.

## GitHub workflow (issues and PRs)

Issue templates: **Feature**, **Bug**, and **Chore**.

| Template    | Use for                                                       |
| ----------- | ------------------------------------------------------------- |
| **Feature** | New capability, infrastructure, or larger improvements        |
| **Bug**     | Broken behavior                                               |
| **Chore**   | Tooling, refactors, deps, maintenance without a product story |

Large work can be a single **Feature** issue with a detailed checklist, or split into several focused issues that reference each other. Use **OpenSpec** for design and task breakdown when the change is non-trivial.

### Suggested labels

Apply **type** and **area** as GitHub labels when opening an issue (not in the issue body):

- `type: feature`, `type: bug`, `type: chore`
- `area: web`, `area: payment`, `area: orders`, `area: cart`, `area: products`, `area: users`, `area: infra`, `area: ci`, `area: observability`, `area: docs`

### Example flow

1. Open a **Feature**, **Bug**, or **Chore** issue using the template.
2. Optional: create a matching **OpenSpec change** for design/tasks.
3. Branch and implement; open a PR using the PR template (`Closes #<issue>`).
4. Merge when CI passes and the issue acceptance criteria are met.

Mapping OpenSpec to GitHub:

| OpenSpec         | GitHub                                            |
| ---------------- | ------------------------------------------------- |
| Change proposal  | Feature issue (optional but recommended)          |
| `tasks.md` items | Checklist in the issue or follow-up issues        |
| PR               | `Closes #<issue>`; OpenSpec change in PR template |

## Continuous integration

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

### Container images (GHCR)

Pushes to `main` that touch image-related paths trigger `.github/workflows/release-images.yml`. The workflow uses the same path-filter fan-out pattern as CI tests: only affected images are built and pushed (for example a `services/web/**` change rebuilds `web`, not all twelve). Changes under `shared/**` or `go.work` fan out to the dependent service images. `workflow_dispatch` builds all images.

- `ghcr.io/<repository>/<image>:<commit-sha>`
- `ghcr.io/<repository>/<image>:main` (rolling tag)

Local Tilt development continues to use unqualified image names (for example `refurbished-marketplace/web`). Helm chart values are not yet wired to GHCR — that is deferred to the ArgoCD GitOps phase.

**Path-filter fan-out for tests:**

| Changed paths                 | Tests triggered                             |
| ----------------------------- | ------------------------------------------- |
| `services/<name>/**`          | That service only                           |
| `shared/proto/**`             | users, products, orders, cart, payment, web |
| `shared/auth/**`              | users, web                                  |
| `shared/messaging/**`         | products, orders, payment                   |
| `shared/testutil/postgres/**` | users, products, orders, payment            |
| `shared/testutil/kafka/**`    | products, orders, payment                   |
| `shared/testutil/redis/**`    | cart                                        |

Local formatting and codegen drift checks (`treefmt`, `generate-proto`, `sqlc-gen`, `templ generate`) stay in devenv/git hooks — they are not run in CI.

## Pull requests

- Use the PR template in `.github/pull_request_template.md`.
- Keep PRs focused; prefer one issue per PR when practical.
- Mention the OpenSpec change name when applicable.
- Describe how you verified the change (usually via Tilt).

## Architecture references

- Service boundaries and stack: [README.md](README.md)
- Order/inventory/payment flow: [docs/order-placement.md](docs/order-placement.md)
- Capability requirements: `openspec/specs/`

## Questions

Open a **Feature**, **Bug**, or **Chore** issue. For multi-step work, use one Feature issue with a checklist or split into several linked issues.
