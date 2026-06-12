# Code generation

Run these from inside `devenv shell`:

| Command          | Purpose                                                                  |
| ---------------- | ------------------------------------------------------------------------ |
| `generate-proto` | Regenerate Go code from `**/proto/*/v1/*.proto`                          |
| `sqlc-gen`       | Regenerate sqlc query code for services with `sqlc.yaml`                 |
| `templ generate` | Regenerate `templ` views (run from `services/web` when templates change) |
| `tidy`           | `go mod tidy` across modules and `go work sync` for `go.work`            |

Edit SQL migrations under `services/<service>/db/migrations/` and queries under `services/<service>/db/queries/`, then run `sqlc-gen` when query shapes change.

## Formatting

Formatting is handled by devenv/git hooks via `treefmt` (`gofumpt`, `sqruff`, etc.). Local formatting drift is not checked in CI.

## Go workspace

Local Go development uses a root `go.work` so tools like gopls and `golangci-lint` see the whole repo. Container builds keep `ENV GOWORK=off` and copy only the service plus required `shared/` paths, so images still build a single module.
