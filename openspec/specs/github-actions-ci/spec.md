# GitHub Actions CI

## Purpose

Define GitHub Actions CI for lint, selective service tests, and vulnerability scanning without publishing images.

## Requirements

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

### Requirement: Weekly vulnerability scanning

The CI workflow SHALL run `govulncheck` for every Go module through `scripts/vulncheck.sh` on the weekly schedule. The scheduled job SHALL fail when vulnerabilities are found; pull requests SHALL NOT be blocked by it.

#### Scenario: Weekly full scan

- **WHEN** CI runs on the scheduled weekly workflow trigger
- **THEN** `govulncheck` scans all Go modules

#### Scenario: Pull requests are not blocked

- **WHEN** CI runs on a pull request or push
- **THEN** no `govulncheck` job runs

### Requirement: Go toolchain version

The repository CI SHALL run lint and test jobs inside the devenv shell, which pins the Go toolchain.

#### Scenario: CI Go version

- **WHEN** CI runs lint or test jobs
- **THEN** they use the Go toolchain provided by devenv

### Requirement: golangci-lint configuration

The repository SHALL generate a root `golangci-lint` configuration from `repo.golangci` in `devenv.nix` that excludes generated code paths consistent with local formatting excludes (protobuf outputs, templ outputs, sqlc database outputs).

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

- `shared/proto/**` → users, products, inventory, orders, cart, web, search
- `shared/auth/**` → users, web
- `shared/messaging/**` → products, inventory, orders, search
- `shared/err/dberr/**` → users, inventory, orders
- `shared/err/grpcerr/**` → users, products, inventory, orders, cart, search
- `shared/runtime/**` → users, products, inventory, orders, cart, web, search
- `shared/observe/log/**` → users, products, inventory, orders, cart, web, search
- `shared/observe/trace/**` → products, inventory, orders, web, search
- `shared/testutil/postgres/**` → users, inventory, orders
- `shared/testutil/kafka/**` → products, inventory, orders, search
- `shared/testutil/redis/**` → cart
- `shared/testutil/mongo/**` → products
- `shared/testutil/meilisearch/**` → search

#### Scenario: Shared proto change

- **WHEN** a pull request modifies files under `shared/proto/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, web, and search

#### Scenario: Shared messaging change

- **WHEN** a pull request modifies files under `shared/messaging/**`
- **THEN** CI runs tests for products, inventory, orders, and search and does not run tests for cart solely due to that change

#### Scenario: Shared observe/trace change

- **WHEN** a pull request modifies files under `shared/observe/trace/**`
- **THEN** CI runs tests for products, inventory, orders, web, and search

#### Scenario: Shared observe/log change

- **WHEN** a pull request modifies files under `shared/observe/log/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, web, and search

#### Scenario: Shared mongo testutil change

- **WHEN** a pull request modifies files under `shared/testutil/mongo/**`
- **THEN** CI runs tests for products and does not run tests for unrelated services solely due to that change

#### Scenario: Shared meilisearch testutil change

- **WHEN** a pull request modifies files under `shared/testutil/meilisearch/**`
- **THEN** CI runs tests for search and does not run tests for unrelated services solely due to that change

### Requirement: Web service tests included

The CI workflow SHALL include `services/web` in selective testing when web paths change or when shared fan-out selects web (proto, auth, runtime, observe/log, or observe/trace changes).

#### Scenario: Web-only change

- **WHEN** a pull request modifies files only under `services/web/**`
- **THEN** CI runs `go test ./...` in `services/web`

### Requirement: Testcontainers-compatible test execution

When a selected service test job runs for a module that uses Testcontainers, CI SHALL execute tests on a GitHub-hosted runner with Docker available.

#### Scenario: Backend service tests with containers

- **WHEN** CI runs tests for users, products, inventory, orders, payment, or cart due to path selection
- **THEN** the test job uses the runner Docker environment sufficient for Testcontainers

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

### Requirement: Inventory participates in service CI

CI SHALL include inventory in the lint module list, service path filters and outputs, test matrix, and vulnerability-scan matrix using the same triggers and shared dependency rules as other Go services.

#### Scenario: Inventory changes are checked

- **WHEN** inventory code or its shared dependencies change in a pull request
- **THEN** inventory lint, tests, and vulnerability scanning SHALL run

#### Scenario: Scheduled vulnerability scan

- **WHEN** the scheduled full vulnerability scan runs
- **THEN** inventory SHALL be scanned alongside the other services

### Requirement: Search participates in service CI

CI SHALL include search in the lint module list, service path filters and outputs, test matrix, and vulnerability-scan matrix using the same triggers and shared dependency rules as other Go services.

#### Scenario: Search changes are checked

- **WHEN** search code or its shared dependencies change in a pull request
- **THEN** search lint, tests, and vulnerability scanning SHALL run

#### Scenario: Scheduled vulnerability scan includes search

- **WHEN** the scheduled full vulnerability scan runs
- **THEN** search SHALL be scanned alongside the other services

### Requirement: Release builds pin production chart image tags

After successful image publication, CI SHALL publish a deployment ref pinning all child Applications to the built Git revision and immutable image tag. It SHALL deploy trusted PRs to dev and main to prod, without writing an image-pin commit back to main or deploying fork PRs.

#### Scenario: Images build on a branch push

- **WHEN** all images build successfully for a commit on a branch
- **THEN** CI deploys the matching immutable Git/image revision through the root without writing an overlay-pin commit

#### Scenario: Branch moved during the build

- **WHEN** the branch head no longer matches the built SHA
- **THEN** CI deploys the matching immutable Git/image revision through the root without writing an overlay-pin commit
