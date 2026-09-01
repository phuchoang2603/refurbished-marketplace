# Code generation

Run these from inside `devenv shell`:

| Command          | Purpose                                                      |
| ---------------- | ------------------------------------------------------------ |
| `generate-proto` | Regenerate Go code from `**/proto/*/v1/*.proto`              |
| `sqlc-gen`       | Regenerate sqlc query code for services with `sqlc.yaml`     |
| `tidy`           | `go work sync` — keeps workspace module dependencies aligned |

Web `templ` and Tailwind: run generators locally and commit the outputs (`*_templ.go`, `static/app.css`). CI builds the `web` image from those files.

Edit SQL migrations under `services/<service>/db/migrations/` and queries under `services/<service>/db/queries/`, then run `sqlc-gen` when query shapes change.

## Formatting

Formatting is handled by devenv/git hooks via `treefmt` (`gofumpt`, `sqruff`, etc.). Local formatting drift is not checked in CI.

## Go workspace

The repo uses a root [`go.work`](../../go.work) file for local development and container builds.

- **`use`**: every service, shared library, and tool module in the monorepo
- **`replace`**: maps local `github.com/phuchoang2603/refurbished-marketplace/shared/...` modules to relative `./shared/...` paths in each consuming `go.mod` (and again in `go.work` for workspace builds)

Shared library layout:

| Path                   | Role                                      |
| ---------------------- | ----------------------------------------- |
| `shared/proto`         | gRPC / protobuf contracts                 |
| `shared/auth`          | JWT and auth config                       |
| `shared/messaging`     | Kafka helpers                             |
| `shared/runtime`       | Service bootstrap (HTTP, gRPC, Postgres…) |
| `shared/observe/trace` | OpenTelemetry tracing                     |
| `shared/observe/log`   | JSON slog + TraceId correlation           |
| `shared/err/dberr`     | DB error / null helpers                   |
| `shared/err/grpcerr`   | gRPC status mapping                       |
| `shared/testutil/*`    | Testcontainers helpers                    |

Run `tidy` (alias for `go work sync`) after changing `go.mod` files. `go get` and `go mod tidy` inside a module directory work because each `go.mod` has `replace` directives for local shared modules.

Build from the repo root:

```bash
go build ./services/users/cmd/users
go test ./services/users/...
```

Container images copy `go.work`, `go.work.sum`, `shared/`, `services/`, and `tools/`, then run `go build` from `/src` so the workspace resolves shared modules the same way as local dev.

Generic templates under `infra/docker/`:

| Dockerfile                                                                  | Purpose                                                              |
| --------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| [`go-service.Dockerfile`](../../infra/docker/go-service.Dockerfile)         | Go microservices and tools (`BUILD_PKG`, `BUILD_BIN`, `EXPOSE_PORT`) |
| [`goose-migrator.Dockerfile`](../../infra/docker/goose-migrator.Dockerfile) | DB migrations (`MIGRATIONS_DIR`)                                     |
| [`web.Dockerfile`](../../infra/docker/web.Dockerfile)                       | Web BFF (same workspace build plus static assets)                    |
