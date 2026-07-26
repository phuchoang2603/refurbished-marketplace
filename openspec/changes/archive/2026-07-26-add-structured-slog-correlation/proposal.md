## Why

Marketplace services still emit unstructured `log.Printf` lines with no `trace_id` / `span_id`, so VictoriaLogs (already deployed) is hard to search and Grafana cannot jump from a span to matching logs. Issue [#2](https://github.com/phuchoang2603/refurbished-marketplace/issues/2) asks for JSON `slog` via `shared/observe/log`, TraceId-based correlation, and Trace → logs wiring now that VL + OTEL tracing are in place.

## What Changes

- Add `shared/observe/log`: JSON-only slog bootstrap, default `service` attr, context helpers that inject `trace_id` / `span_id` from the active OTEL span, and light redaction helpers.
- Wire bootstrap through `shared/runtime` (same pattern as tracing); migrate production-path `log.Printf` / `log.Fatal*` call sites to slog.
- **web:** replace chi `middleware.Logger` with slog request logging; **remove** unused `middleware.RequestID` (TraceId is the join key).
- **gRPC** and **Kafka consumers:** structured access / handle logs with method or topic/partition/offset plus `trace_id` when a span is active.
- **Domain hot paths (checkout):** structured Info/Warn with `order_id` and related IDs/outcomes on order create/status, inventory reserve/settle, and payment session/tx/webhook (no card/password payloads).
- Grafana: configure VictoriaTraces (Tempo) **tracesToLogs** against VictoriaLogs; document field conventions and LogSQL examples in `docs/observability.md`.

Non-goals: text log formats; dual RequestID scheme; domain fields on every CRUD line; k6 / load tooling; per-service Prometheus `/metrics`; further shared package moves (already done in #27).

## Capabilities

### New Capabilities

- `structured-logging`: JSON slog bootstrap under `shared/observe/log`, OTEL correlation fields, web/gRPC/Kafka structured access logs, checkout domain hot-path fields, and Trace → logs documentation conventions.

### Modified Capabilities

- `platform-observability`: Require VictoriaTraces datasource Trace → logs correlation to VictoriaLogs; acknowledge application structured logging into the existing VL pipeline (stdout → VLAgent).
- `web`: Replace chi text request logger / RequestID middleware with slog JSON request logs correlated by TraceId.

## Impact

- Touches `shared/observe/log` (new), `shared/runtime`, `shared/messaging`, service `main`/`kafka` entrypoints, `services/web` router, gRPC server startup paths, `infra/charts/observability` Grafana Tempo datasource config, and `docs/observability.md`.
- Depends on VictoriaLogs + VLAgent (#1), OTEL tracing (#3 / PR #24), and `shared/observe/` layout (#27).
- Does not change business protobufs, browser UX, or VLAgent scrape targets.
