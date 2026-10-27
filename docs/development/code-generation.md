# Code generation

Run tasks explicitly via `devenv tasks run <task>`. Tasks use `execIfModified` to skip unchanged inputs and do not run on shell entry:

| Task / Command                      | Purpose                                                     |
| ----------------------------------- | ----------------------------------------------------------- |
| `devenv tasks run codegen:proto`    | Regenerate Go code from `**/proto/*/v1/*.proto`             |
| `devenv tasks run codegen:sqlc`     | Regenerate sqlc query code for services with `sqlc.yaml`    |
| `devenv tasks run go:tidy`          | `go work sync` — keep workspace module dependencies aligned |
| `devenv tasks run codegen:templ`    | `templ generate` in `services/web`                          |
| `devenv tasks run codegen:tailwind` | Tailwind → `services/web/static/app.css`                    |

Commit `*_templ.go` and `static/app.css`. CI builds the `web` image from those files. For a live watch while editing, run `templ generate --watch` / `tailwindcss … --watch=always` in `services/web`.

Edit SQL migrations under `services/<service>/db/migrations/` and queries under `services/<service>/db/queries/`, then run `devenv tasks run codegen:sqlc` when query shapes change. Products uses the same SQL migrations and sqlc flow through `services/products/internal/catalog`. Search has no SQL store.

## Formatting

Formatting is handled by the commit hook via `treefmt` (`gofumpt`, `sqruff`, etc.) and does not run on shell entry. Run `treefmt` explicitly when needed. Local formatting drift is not checked in CI.

## Go workspace

The repo uses a root [`go.work`](../../go.work) file for local development and container builds.

- **`use`**: every service, shared library, and tool module in the monorepo
- **`replace`**: maps local `github.com/phuchoang2603/refurbished-marketplace/shared/...` modules to relative `./shared/...` paths in each consuming `go.mod` (and again in `go.work` for workspace builds)

Shared library layout:

| Path                    | Role                                       |
| ----------------------- | ------------------------------------------ |
| `shared/proto`          | gRPC / protobuf contracts                  |
| `shared/auth`           | JWT and auth config                        |
| `shared/messaging`      | Kafka helpers                              |
| `shared/runtime`        | Service bootstrap (HTTP, gRPC, Postgres)   |
| `shared/observe/trace`  | OpenTelemetry tracing                      |
| `shared/observe/log`    | JSON slog + TraceId correlation            |
| `shared/observe/metric` | OpenTelemetry RED metrics pushed over OTLP |
| `shared/err/dberr`      | DB error / null helpers                    |
| `shared/err/grpcerr`    | gRPC status mapping                        |
| `shared/testutil/*`     | Testcontainers helpers                     |

Run `devenv tasks run go:tidy` (or `go work sync`) after changing `go.mod` files. `go get` and `go mod tidy` inside a module directory work because each `go.mod` has `replace` directives for local shared modules.

Build from the repo root:

```bash
go build ./services/users/cmd/users
go test ./services/users/...
```

Container images copy `go.work`, `go.work.sum`, `shared/`, `services/`, and `tools/`, then run `go build` from `/src` so the workspace resolves shared modules the same way as local dev.

Generic templates under `infra/docker/`:

| Dockerfile                                                          | Purpose                                                              |
| ------------------------------------------------------------------- | -------------------------------------------------------------------- |
| [`go-service.Dockerfile`](../../infra/docker/go-service.Dockerfile) | Go microservices and tools (`BUILD_PKG`, `BUILD_BIN`, `EXPOSE_PORT`) |
| [`web.Dockerfile`](../../infra/docker/web.Dockerfile)               | Web BFF (same workspace build plus static assets)                    |
