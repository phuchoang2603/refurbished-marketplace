## ADDED Requirements

### Requirement: CI runs on pull requests and main

The repository SHALL run GitHub Actions CI on every pull request and on every push to the `main` branch.

#### Scenario: Pull request opened

- **WHEN** a pull request is opened or updated against the repository
- **THEN** the CI workflow is triggered

#### Scenario: Push to main

- **WHEN** a commit is pushed to the `main` branch
- **THEN** the CI workflow is triggered

### Requirement: Lint all Go modules

The CI workflow SHALL lint every Go module in the repository on each run using `golangci-lint`. The lint job SHALL fail when `golangci-lint` reports errors.

#### Scenario: Lint job succeeds

- **WHEN** CI runs and all Go modules pass `golangci-lint`
- **THEN** the lint job reports success

#### Scenario: Lint job fails on golangci violation

- **WHEN** CI runs and a Go module fails `golangci-lint`
- **THEN** the lint job reports failure and the workflow fails

### Requirement: Lint posts PR review comments via Reviewdog

On pull requests, the CI lint job SHALL post review comments for `golangci-lint` findings using the official Reviewdog GitHub Action, without custom JavaScript comment scripts.

#### Scenario: Pull request lint comments

- **WHEN** CI runs on a pull request and `golangci-lint` reports findings on changed lines
- **THEN** Reviewdog posts PR review comments for those findings

#### Scenario: Push to main uses check annotations

- **WHEN** CI lint runs on a push to `main`
- **THEN** Reviewdog reports `golangci-lint` findings via GitHub Checks rather than PR review comments

### Requirement: Vulnerability scanning via govulncheck SARIF

The CI workflow SHALL run `govulncheck` for service modules selected by the same path filters as integration tests, produce SARIF output, and upload results via `github/codeql-action/upload-sarif`. The job SHALL NOT fail when vulnerabilities are found.

#### Scenario: SARIF uploaded for affected services

- **WHEN** CI runs and path filters select a service module
- **THEN** `govulncheck` scans that service and uploads SARIF to GitHub Code Scanning

#### Scenario: Vulnerabilities do not fail CI

- **WHEN** `govulncheck` reports vulnerabilities
- **THEN** the govulncheck job completes successfully

#### Scenario: Weekly full scan

- **WHEN** CI runs on the scheduled weekly workflow trigger
- **THEN** `govulncheck` scans all service modules regardless of path filters

### Requirement: Go toolchain version

The repository CI and Go module toolchain SHALL use Go **1.26.2**.

#### Scenario: CI Go version

- **WHEN** CI installs Go for lint or test jobs
- **THEN** it uses Go 1.26.2

### Requirement: golangci-lint configuration

The repository SHALL include a root `golangci-lint` configuration that excludes generated code paths consistent with local formatting excludes (protobuf outputs, templ outputs, sqlc database outputs).

#### Scenario: Generated code excluded

- **WHEN** CI runs `golangci-lint` against modules containing generated `*_templ.go` or sqlc-generated database code
- **THEN** those generated paths are not reported as lint violations solely for being generated artifacts

### Requirement: Selective service integration tests

The CI workflow SHALL run `go test ./...` for a service module only when that module is selected by path filters for the change.

#### Scenario: Single service change

- **WHEN** a pull request modifies files only under `services/users/**`
- **THEN** CI runs tests for the users module and does not run tests for unrelated service modules such as cart or web unless shared fan-out applies

#### Scenario: No testable service changes

- **WHEN** a pull request modifies only documentation or non-service paths outside shared fan-out rules
- **THEN** CI skips service test jobs while still running lint

### Requirement: Shared dependency test fan-out

The CI workflow SHALL expand path filters so changes under shared modules trigger tests for dependent service modules according to this map:

- `shared/proto/**` → users, products, orders, cart, payment, web
- `shared/auth/**` → users, web
- `shared/messaging/**` → products, orders, payment
- `shared/testutil/postgres/**` → users, products, orders, payment
- `shared/testutil/kafka/**` → products, orders, payment
- `shared/testutil/redis/**` → cart

#### Scenario: Shared proto change

- **WHEN** a pull request modifies files under `shared/proto/**`
- **THEN** CI runs tests for users, products, orders, cart, payment, and web

#### Scenario: Shared messaging change

- **WHEN** a pull request modifies files under `shared/messaging/**`
- **THEN** CI runs tests for products, orders, and payment and does not run tests for cart solely due to that change

### Requirement: Web service tests included

The CI workflow SHALL include `services/web` in selective testing when web paths change or when shared fan-out selects web (proto or auth changes).

#### Scenario: Web-only change

- **WHEN** a pull request modifies files only under `services/web/**`
- **THEN** CI runs `go test ./...` in `services/web`

### Requirement: Testcontainers-compatible test execution

When a selected service test job runs for a module that uses Testcontainers, CI SHALL execute tests on a GitHub-hosted runner with Docker available.

#### Scenario: Backend service tests with containers

- **WHEN** CI runs tests for users, products, orders, payment, or cart due to path selection
- **THEN** the test job uses the runner Docker environment sufficient for Testcontainers without Colima-specific configuration

### Requirement: Helm validation on chart changes

When a change modifies files under `infra/charts/**`, CI SHALL validate Helm charts by running `helm lint`, rendering manifests with `helm template`, and validating rendered YAML with `kubeconform` using `-ignore-missing-schemas`.

#### Scenario: Chart change triggers helm job

- **WHEN** a pull request modifies files under `infra/charts/**`
- **THEN** CI runs helm lint, helm template, and kubeconform validation for the affected charts

#### Scenario: Non-chart change skips helm job

- **WHEN** a pull request does not modify files under `infra/charts/**`
- **THEN** CI skips the Helm validation job

### Requirement: CI excludes local-only quality gates

The CI workflow SHALL NOT run treefmt, proto generation drift checks, sqlc generation drift checks, or templ generation drift checks.

#### Scenario: Formatter-only local workflow

- **WHEN** CI runs on a pull request
- **THEN** CI does not invoke treefmt or codegen drift verification steps

### Requirement: CI excludes container image publishing

The CI workflow (`ci.yml`) SHALL NOT build or push container images to a registry. Image publishing SHALL be handled by the separate GHCR release workflow.

#### Scenario: CI workflow on main

- **WHEN** CI runs on a push to `main`
- **THEN** `ci.yml` does not publish container images to GHCR

#### Scenario: Merge to main without image workflow trigger

- **WHEN** a commit is merged to `main` without image-related path changes
- **THEN** neither `ci.yml` nor the release workflow publishes images
