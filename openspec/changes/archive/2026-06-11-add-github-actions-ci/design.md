## Context

The repository uses a Go multi-module layout (`services/*`, `shared/*`, `tools/*`) with no `.github/workflows/` today. Local development relies on devenv (treefmt, codegen scripts, Tilt). Integration tests live under `services/<service>/tests/` and use Testcontainers for Postgres, Kafka, and Redis in backend services; `services/web` tests use httptest and fakes without Docker.

Issue #4 scopes broader CI/CD (GHCR image push, ArgoCD). This change implements the **first PR slice**: lint, selective integration tests, and Helm validation only.

## Goals / Non-Goals

**Goals:**

- Run CI on every pull request and push to `main`.
- Lint all Go modules with `golangci-lint`, `go vet`, and `go build`.
- Run `go test ./...` only for service modules affected by changed paths, including shared dependency fan-out.
- Validate Helm charts when `infra/charts/**` changes.
- Keep CI aligned with local tooling versions (Go 1.26.1).

**Non-Goals:**

- GHCR image build/push (follow-up PR).
- treefmt, proto/sqlc/templ drift checks in CI.
- ArgoCD, deploy docs, migration promotion.
- CI for `infra/k8s/*.yaml` raw manifests.
- devenv/Nix in GitHub Actions (plain `setup-go` + action-installed linters is sufficient).

## Decisions

### Single workflow file: `.github/workflows/ci.yml`

One workflow with parallel jobs keeps required checks simple for branch protection (`lint`, optional per-service tests, optional `helm`).

**Alternatives considered:** separate workflows per concern — rejected as unnecessary for current scale.

### Lint job runs on every CI invocation

The lint job loops all Go module roots and runs:

1. `golangci-lint run ./...`
2. `go vet ./...`
3. `go build ./...`

Configuration lives at repo root (`.golangci.yml`) with `run:` paths or a matrix/loop over module directories. `issues:`/`run:` should exclude generated paths already ignored locally (`**/proto/*.go`, `*_templ.go`, `**/database/*.go` where applicable).

**Rationale:** lint is fast and needs no Docker; it catches cross-module breakage when `shared/` changes without running every Testcontainers suite.

### Path-filtered tests via `dorny/paths-filter@v3`

A `changes` job emits boolean outputs per testable service. Each service has a dedicated job:

```yaml
test-users:
  needs: changes
  if: needs.changes.outputs.users == 'true'
```

Test command: `go test ./...` in the module directory (e.g. `services/users`).

Docker: default Ubuntu runner socket is sufficient for Testcontainers; no Colima-specific env vars from devenv.

### Shared dependency fan-out (Option B)

Service-local paths trigger only that service. Shared paths fan out to dependent services:

| Changed paths                 | Trigger tests for                           |
| ----------------------------- | ------------------------------------------- |
| `services/users/**`           | users                                       |
| `services/products/**`        | products                                    |
| `services/orders/**`          | orders                                      |
| `services/payment/**`         | payment                                     |
| `services/cart/**`            | cart                                        |
| `services/web/**`             | web                                         |
| `shared/proto/**`             | users, products, orders, cart, payment, web |
| `shared/auth/**`              | users, web                                  |
| `shared/messaging/**`         | products, orders, payment                   |
| `shared/testutil/postgres/**` | users, products, orders, payment            |
| `shared/testutil/kafka/**`    | products, orders, payment                   |
| `shared/testutil/redis/**`    | cart                                        |

**Rationale:** avoids running all six service suites on unrelated edits while still retesting consumers when shared contracts or test helpers change.

**Alternatives considered:** strict module-only (no fan-out) — rejected as too risky for `shared/proto` edits.

### Include `services/web` tests

Web tests run when web paths or fan-out from `shared/proto` / `shared/auth` fire. They do not require Docker but share the same selective pattern for consistency.

### Helm job gated on chart paths

The `helm` job runs only when `infra/charts/**` changes:

1. `helm lint` on `infra/charts/refurbished-marketplace` and `infra/charts/kafka`
2. `helm template` both charts into temp manifests (fixed `--namespace ecommerce`)
3. `kubeconform -summary -ignore-missing-schemas` on rendered YAML

**Rationale:** charts render Strimzi and CNPG CRDs without bundled schemas; `-ignore-missing-schemas` still validates standard Kubernetes kinds (Deployment, Service, Job, etc.).

Install tools in-job: `helm` (e.g. `azure/setup-helm`), `kubeconform` (binary download or container action).

**Alternatives considered:** always run helm — rejected per scope preference; pin CRD OpenAPI schemas — deferred as follow-up hardening.

### golangci-lint configuration

Add `.golangci.yml` at repo root with a conservative default set (e.g. `govet`, `staticcheck`, `errcheck`, `unused`, `gosimple`, ` ineffassign`) and exclusions for generated code matching devenv treefmt excludes.

Run via `golangci/golangci-lint-action` with version pinned; execute per module directory because there is no `go.work`.

### Workflow triggers

```yaml
on:
  pull_request:
  push:
    branches: [main]
```

No path filter at workflow level — lint always runs; tests and helm self-gate with `if:`.

### Required checks strategy

Document in CONTRIBUTING (optional follow-up): branch protection should require `lint`. Test and helm jobs may be skipped on path-no-op PRs — acceptable; optional aggregate job not required for v1.

## Risks / Trade-offs

- **[Missed integration coverage on tangential edits]** → Mitigated by shared fan-out map; full suite still runnable locally via `devenv shell`.
- **[kubeconform skips CRD validation]** → Mitigated by `helm lint` + template smoke; CRD schema pinning is a future improvement.
- **[golangci-lint false positives on generated code]** → Mitigate with explicit excludes in `.golangci.yml`.
- **[Go 1.26.1 availability on `setup-go`]** → Pin exact version; fall back to `go-version-file` from a representative `go.mod` if needed.
- **[Path filter drift when modules/dependencies change]** → Document fan-out table in design; update filters when new services or shared modules appear.

## Migration Plan

1. Add `.golangci.yml` and fix any lint findings blocking CI (or tune config minimally).
2. Add `.github/workflows/ci.yml`.
3. Open a test PR touching one service — confirm only that service's test job runs.
4. Open a test PR touching `shared/proto` — confirm fan-out jobs run.
5. Open a test PR touching `infra/charts/**` — confirm helm job runs.
6. Enable branch protection requiring `lint` after first green `main` run.

Rollback: delete workflow and golangci config; no runtime impact.

## Open Questions

- None blocking implementation; CRD schema pinning for strict kubeconform is optional follow-up.
