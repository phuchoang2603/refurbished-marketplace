## 1. Shared slog package

- [x] 1.1 Add `shared/observe/log` module (JSON slog handler to stdout, default `service` attr, context helpers for `trace_id` / `span_id`, light redaction helpers)
- [x] 1.2 Register the module in `go.work`, CI `GO_MODULE_GLOBS` / path filters, and docs/OpenSpec shared-layout notes as needed
- [x] 1.3 Add `shared/runtime` logging bootstrap (init early in service `main`s) and migrate runtime `log.Printf` call sites

## 2. Access logs: web, gRPC, Kafka

- [x] 2.1 Replace chi `middleware.Logger` with slog HTTP access logging; remove `middleware.RequestID` from `services/web/cmd/web/router.go`
- [x] 2.2 Add structured gRPC unary access logging in `shared/runtime` (compose with `sharedtrace.GRPCServerOptions()`; call `shared/observe/log` helpers — do not put interceptor in observe modules)
- [x] 2.3 Emit structured Kafka consumer error (and startup) logs with `topic` / `partition` / `offset` / `trace_id` in `shared/messaging`
- [x] 2.4 Migrate remaining service `main` / `kafka` entrypoint `log.Printf` / `log.Fatal*` production paths to slog

## 3. Grafana Trace → logs and docs

- [x] 3.1 Configure VictoriaTraces (`uid: VictoriaTraces`) `jsonData.tracesToLogsV2` → `datasourceUid: VictoriaLogs` with `filterByTraceID: true` in observability chart values (local + staging); fall back to VL customQuery if needed
- [x] 3.2 Document field conventions, LogSQL examples, Trace → logs steps, and redaction notes in `docs/observability.md`

## 3b. Domain hot-path fields

- [x] 3.3 Emit structured domain logs on checkout hot paths (orders create/status, inventory reserve/settle, payment session/tx/webhook) with `order_id` and related IDs/outcomes
- [x] 3.4 Update issue #2, OpenSpec structured-logging spec, and docs for domain fields

## 4. Verify

- [x] 4.1 `tilt up`: generate web + gRPC + checkout traffic; confirm JSON lines reach VictoriaLogs
- [x] 4.2 In Grafana Explore (VL), filter by `service` and a known checkout `trace_id` from VictoriaTraces
- [x] 4.3 From a span, use Trace → logs and land on matching log lines; spot-check payment/callback paths for no full sensitive payloads
