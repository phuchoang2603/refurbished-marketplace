## 1. Lint configuration

- [x] 1.1 Add root `.golangci.yml` with linters (`govet`, `staticcheck`, `errcheck`, `unused`, `gosimple`, `ineffassign`) and excludes for generated paths (`**/*_templ.go`, `**/proto/*.go`, `**/database/*.go`)
- [x] 1.2 Run `golangci-lint` locally across all Go module roots and fix or tune config for any blocking findings

## 2. CI workflow scaffold

- [x] 2.1 Create `.github/workflows/ci.yml` with triggers on `pull_request` and push to `main`
- [x] 2.2 Add a `changes` job using `dorny/paths-filter@v3` with service filters and shared fan-out paths per design (proto → six services; auth → users/web; messaging → products/orders/payment; testutil postgres/kafka/redis → dependent services)
- [x] 2.3 Add a `lint` job on `ubuntu-latest` with `actions/setup-go` (Go 1.26.1) that loops all module directories and runs `golangci-lint run ./...`, `go vet ./...`, and `go build ./...`

## 3. Selective integration tests

- [x] 3.1 Add `test-users`, `test-products`, `test-orders`, `test-payment`, and `test-cart` jobs gated by `changes` outputs; each runs `go test ./...` in its service directory on the default runner Docker socket
- [x] 3.2 Add `test-web` job gated by `changes` output for web (including fan-out from `shared/proto` and `shared/auth`); runs `go test ./...` in `services/web`

## 4. Helm validation

- [x] 4.1 Add `helm` job gated on `infra/charts/**` path filter with `azure/setup-helm` (or equivalent)
- [x] 4.2 Run `helm lint` and `helm template` for `infra/charts/refurbished-marketplace` and `infra/charts/kafka` (namespace `ecommerce`)
- [x] 4.3 Install `kubeconform` and validate rendered manifests with `-summary -ignore-missing-schemas`

## 5. Verification

- [x] 5.1 Validate workflow YAML structure (e.g. `actionlint` locally if available, or manual review)
- [x] 5.2 Dry-run path filter expectations against sample change sets (service-only, shared/proto, charts-only)
- [x] 5.3 Document expected branch protection check: require `lint` job; note test/helm jobs may skip on path-no-op PRs
