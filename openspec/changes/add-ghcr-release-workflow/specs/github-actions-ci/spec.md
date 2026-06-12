## MODIFIED Requirements

### Requirement: Lint all Go modules

The CI workflow SHALL lint every Go module in the repository on each run using `golangci-lint` and informational `govulncheck`. The lint job SHALL fail when `golangci-lint` reports errors and SHALL NOT fail when `govulncheck` reports vulnerabilities.

#### Scenario: Lint job succeeds

- **WHEN** CI runs and all Go modules pass `golangci-lint`
- **THEN** the lint job reports success regardless of `govulncheck` findings

#### Scenario: Lint job fails on golangci violation

- **WHEN** CI runs and a Go module fails `golangci-lint`
- **THEN** the lint job reports failure and the workflow fails

#### Scenario: Lint job succeeds despite govulncheck findings

- **WHEN** CI runs and `govulncheck` reports vulnerabilities but `golangci-lint` passes
- **THEN** the lint job reports success

### Requirement: CI excludes container image publishing

The CI workflow (`ci.yml`) SHALL NOT build or push container images to a registry. Image publishing SHALL be handled by the separate GHCR release workflow.

#### Scenario: CI workflow on main

- **WHEN** CI runs on a push to `main`
- **THEN** `ci.yml` does not publish container images to GHCR

#### Scenario: Merge to main without image workflow trigger

- **WHEN** a commit is merged to `main` without image-related path changes
- **THEN** neither `ci.yml` nor the release workflow publishes images

## ADDED Requirements

### Requirement: Lint posts PR review comments via Reviewdog

On pull requests, the CI lint job SHALL post review comments for `golangci-lint` and `govulncheck` findings using official Reviewdog GitHub Actions, without custom JavaScript comment scripts.

#### Scenario: Pull request lint comments

- **WHEN** CI runs on a pull request and lint tools report findings on changed lines
- **THEN** Reviewdog posts PR review comments for those findings

#### Scenario: Push to main uses check annotations

- **WHEN** CI lint runs on a push to `main`
- **THEN** Reviewdog reports findings via GitHub Checks rather than PR review comments

### Requirement: Go toolchain version

The repository CI and Go module toolchain SHALL use Go **1.26.2**.

#### Scenario: CI Go version

- **WHEN** CI installs Go for lint or test jobs
- **THEN** it uses Go 1.26.2
