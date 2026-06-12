## Context

CI (`ci.yml`) runs lint (golangci-lint + govulncheck), path-filtered tests, and Helm validation. Images under `infra/docker/` are built by Tilt using local names like `refurbished-marketplace/web`. Helm values use unqualified image names for local clusters.

Issue #4 next step: publish images to GHCR on `main` without ArgoCD wiring yet. The prior `add-github-actions-ci` change explicitly deferred image push and pinned Go 1.26.2 before a later bump to 1.26.4 for `govulncheck` noise; this change reverts to **1.26.2** and treats vulnerabilities as informational in CI.

## Goals / Non-Goals

**Goals:**

- Push all 12 Dockerfiles in `infra/docker/` to GHCR on relevant `main` pushes.
- Use standard GitHub Actions (`docker/login-action`, `docker/build-push-action`, `reviewdog/action-golangci-lint`, `reviewdog/action-setup`).
- Post PR review comments for golangci-lint (blocking) and govulncheck (informational).
- Path-filter the release workflow so doc-only merges do not rebuild images.
- Revert Go to 1.26.2 repo-wide.

**Non-Goals:**

- ArgoCD, `values-staging.yaml`, `imagePullSecrets`, or chart template registry helpers.
- PR build-only image workflows (push only on `main`).
- Custom `actions/github-script` comment bots.
- Failing CI on `govulncheck` findings.

## Decisions

### Separate workflow: `release-images.yml`

Keep image build/push out of `ci.yml` for speed and clearer permissions (`packages: write` only on release workflow).

**Trigger:**

```yaml
on:
  push:
    branches: [main]
    paths:
      - infra/docker/**
      - services/**
      - shared/**
      - tools/**
      - go.work
      - "**/go.mod"
  workflow_dispatch:
```

When triggered, build and push **all 12 images** (shared/`go.work` changes can affect every service image).

**Alternatives considered:** Per-image path filters in matrix — more precise but high maintenance; defer unless build times become painful.

### GHCR naming and tags

- Repository: `ghcr.io/${{ github.repository }}/<image>`
- Tags per image:
  - `${{ github.sha }}` (immutable)
  - `main` (rolling, optional convenience)

Example: `ghcr.io/phuchoang2603/refurbished-marketplace/web:abc123`

Matrix `include` lists all 12 images aligned with Tilt/Dockerfiles:

| image                     | dockerfile                                        |
| ------------------------- | ------------------------------------------------- |
| web                       | infra/docker/web.Dockerfile                       |
| users                     | infra/docker/users.Dockerfile                     |
| users-migrator            | infra/docker/users-migrator.Dockerfile            |
| products                  | infra/docker/products.Dockerfile                  |
| products-migrator         | infra/docker/products-migrator.Dockerfile         |
| orders                    | infra/docker/orders.Dockerfile                    |
| orders-migrator           | infra/docker/orders-migrator.Dockerfile           |
| cart                      | infra/docker/cart.Dockerfile                      |
| payment                   | infra/docker/payment.Dockerfile                   |
| payment-migrator          | infra/docker/payment-migrator.Dockerfile          |
| payment-gateway-simulator | infra/docker/payment-gateway-simulator.Dockerfile |
| connect-debezium          | infra/docker/connect-debezium.Dockerfile          |

Each job:

```yaml
- uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- uses: docker/build-push-action@v6
  with:
    context: .
    file: ${{ matrix.dockerfile }}
    push: true
    tags: |
      ghcr.io/${{ github.repository }}/${{ matrix.image }}:${{ github.sha }}
      ghcr.io/${{ github.repository }}/${{ matrix.image }}:main
```

Build context remains repo root `.`; Dockerfiles keep `ENV GOWORK=off`.

### Lint PR comments via Reviewdog (no custom JS)

**golangci-lint (blocking):**

```yaml
permissions:
  contents: read
  pull-requests: write
  checks: write

- uses: actions/checkout@v4
  with:
    fetch-depth: 0

- uses: reviewdog/action-golangci-lint@v2
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    go_version: "1.26.2"
    golangci_lint_version: v2.12.2
    golangci_lint_flags: --config=.golangci.yml ${{ env.GO_MODULE_GLOBS }}
    reporter: ${{ github.event_name == 'pull_request' && 'github-pr-review' || 'github-check' }}
    filter_mode: diff_context
    fail_level: error
```

Uses the official action; no manual install or github-script summary.

**govulncheck (informational, never fails lint):**

Run per module (same as today — `govulncheck` requires a module directory). Use `reviewdog/action-setup@v1` plus Reviewdog CLI with `-fail-level=none`:

```yaml
- uses: reviewdog/action-setup@v1

- name: govulncheck
  continue-on-error: true
  if: always()
  env:
    REVIEWDOG_GITHUB_API_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    for dir in $(grep -E '^\s*\./' go.work | tr -d ' \t'); do
      (cd "$dir" && govulncheck ./...) | \
        reviewdog -f=govulncheck -name=govulncheck \
          -reporter=${{ github.event_name == 'pull_request' && 'github-pr-review' || 'github-check' }} \
          -filter-mode=diff_context -fail-level=none -level=warning || true
    done
```

`continue-on-error: true` + `fail-level=none` ensures vulnerabilities are reported but **never fail** the job. Install `govulncheck` via `go install` in a prior step (Reviewdog action does not install it).

**Alternatives considered:** Failing lint on govulncheck — rejected (user does not want repeated Go bump pressure). Custom JS sticky comments — rejected.

### Go version revert to 1.26.2

Update `go.work`, all `go.mod`, `GO_VERSION` in workflows, and `golang:*-alpine` builder stages in service Dockerfiles. Run `go work sync` via `tidy` after bump.

### Keep `ci.yml` separate from `release-images.yml`

Lint/test/helm permissions stay minimal (`contents: read` + PR write only on lint job if split, or workflow-level for lint comments).

## Risks / Trade-offs

- **[Review comments only on diff lines]** → `filter_mode: diff_context` limits inline comments; full findings remain in job logs.
- **[All 12 images rebuild on shared change]** → Correct for dependency safety; costs runner minutes.
- **[Private GHCR packages]** → Clusters need `imagePullSecrets` later (ArgoCD phase).
- **[govulncheck noise without enforcement]** → Acceptable trade-off; developers see comments without blocked merges for stdlib CVEs.

## Migration Plan

1. Revert Go to 1.26.2 across repo.
2. Update `ci.yml` lint to Reviewdog-based golangci + informational govulncheck.
3. Add `release-images.yml` and merge a change touching `services/**` to verify GHCR packages.
4. Confirm packages visible under GitHub **Packages** for the repository.

Rollback: delete/disable `release-images.yml`; revert lint job changes.

## Open Questions

- None blocking implementation.
