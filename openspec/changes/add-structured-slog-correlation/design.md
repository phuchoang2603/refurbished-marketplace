## Context

VictoriaLogs + VLAgent already scrape `ecommerce` / `kafka` stdout (#1). Distributed tracing exports OTEL spans with W3C TraceIds (#3). Shared layout places observability under `shared/observe/` (#27). Services still use stdlib `log` text lines (mostly `shared/runtime` and service `main`/`kafka` entrypoints); web uses chi `middleware.Logger` + unused `middleware.RequestID`.

Issue [#2](https://github.com/phuchoang2603/refurbished-marketplace/issues/2) closes the Trace → logs gap with JSON slog and Grafana correlation.

## Goals / Non-Goals

**Goals:**

- JSON-only slog via `shared/observe/log`, correlated with active OTEL spans (`trace_id`, `span_id`).
- Consistent structured access logs for web HTTP, gRPC unary, and Kafka consumer handle/error paths.
- Grafana Trace → logs from VictoriaTraces (Tempo) to VictoriaLogs using `trace_id`.
- Document field conventions and LogSQL examples in `docs/observability.md`.

**Non-Goals:**

- Text / multi-format log handlers.
- Chi RequestID or any second request-id scheme.
- Logging full payment/card/password payloads.
- Domain IDs on every log line (optional later).
- Per-service Prometheus `/metrics` or k6 tooling.
- Changing VLAgent scrape namespaces or VL retention.

## Decisions

### Decision: `shared/observe/log` module next to `shared/observe/trace`

New Go workspace module `refurbished-marketplace/shared/observe/log` with slog JSON handler to stdout, default attrs via `Logger.With`, and context-aware helpers.

Follow Go slog practices ([Getting started with slog](https://go.dev/blog/slog), Go 1.26 `log/slog` API):

- `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{…})` then `slog.SetDefault` so package-level `Info`/`Error`/`InfoContext` work.
- Prefer `InfoContext` / `ErrorContext` / `LogAttrs` so context is available for TraceId extraction; use `slog.Attr` / `LogAttrs` on hot paths to cut allocations.
- Attach stable attrs with `logger.With("service", name)` once at bootstrap — not per call.
- Use `HandlerOptions.ReplaceAttr` for cross-cutting transforms (e.g. rename/omit keys, normalize `time`/`level` if VL parsing needs it). Keep `AddSource` off by default unless debugging.
- Optional `slog.LevelVar` if we later need runtime level changes; v1 can stay Info+.
- Redaction: prefer `LogValuer` on sensitive types and/or `ReplaceAttr` denylist — do not invent a full PII framework.
- Bridge leftovers: `slog.NewLogLogger(handler, level)` only if a dependency still needs `*log.Logger`.

Correlation helpers (thin, in-module — no slog-context dependency for v1):

- On each `*Context` log, if `trace.SpanFromContext(ctx).SpanContext().IsValid()`, add snake_case `trace_id` / `span_id` as hex (matches Tempo / W3C). Reference pattern: community `slog-context` + OTEL extractors; we implement the same idea with `otel/trace` only so `shared/observe/log` does not pull handler middleware deps.
- Do **not** store `*slog.Logger` in context as the primary pattern; extract IDs at log time from the span on `ctx`.

Wire `InitLogging(serviceName)` from `shared/runtime` early in each `main`, before other bootstrap logs. Replace `log.Printf` / `log.Fatal*` on production paths with slog (`Error`/`Info` + `os.Exit` where fatal).

**Alternatives considered:** zap/zerolog (extra deps); put logging inside `shared/observe/trace` (wrong cohesion); adopt `veqryn/slog-context` (nice DX, but avoid new deps for v1); OTEL Logs SDK export (heavier; VLAgent already scrapes stdout).

### Decision: JSON only, field names for Grafana join

Emit snake_case attrs: `service`, `trace_id`, `span_id`, plus access-log fields (`method`, `path`, `status`, `duration_ms`, gRPC `grpc_method` / `grpc_code`, Kafka `topic` / `partition` / `offset`). Use hex TraceId/SpanId strings matching OTEL / Tempo so Explore Trace → logs and LogSQL filters join on `trace_id`.

Avoid slog `Group` for correlation fields (grouping nests JSON and complicates VL field filters). Groups are fine later for optional nested payloads.

**Alternatives considered:** dual text+JSON via `LOG_FORMAT` (ticket forbids text); OpenTelemetry Logs SDK export (heavier; VLAgent already collects stdout).

### Decision: Retire chi RequestID; TraceId is the join key

Remove `middleware.RequestID` from `services/web/cmd/web/router.go`. Replace `middleware.Logger` with a small slog middleware that logs after the response with status, duration, and context-derived `trace_id` (from `otelhttp` span on the request). Use `InfoContext(r.Context(), …)`.

**Alternatives considered:** Keep RequestID and also log it (ticket explicitly out of scope); log X-Request-Id only (weaker than TraceId for e2e).

### Decision: gRPC interceptor in `shared/runtime` composing log helpers

Register a unary server interceptor in **`shared/runtime`** (alongside existing `sharedtrace.GRPCServerOptions()` composition), calling `shared/observe/log` helpers for method, code, duration, and context `trace_id`.

Keep interceptors out of `shared/observe/trace` and out of `shared/observe/log` to avoid circular module deps (`trace` must not import `log`; `log` must not import gRPC server wiring).

- **Kafka:** in `shared/messaging` consumer loop, structured Error logs on handler failure MUST include topic, partition, offset, `trace_id`. Startup “consumer started” lines migrate to slog Info. Avoid Info-per-success unless needed.

**Alternatives considered:** interceptor inside `observe/trace` (forces trace→log or log→grpc coupling); per-service interceptors (duplication); log every successful Kafka message at Info (noisy).

### Decision: Grafana tracesToLogsV2 with cluster-verified UIDs

Inspected staging Grafana provisioning (`monitoring` / `vmks-grafana-ds` and pod `/etc/grafana/provisioning/datasources/datasource.yaml` via `dev-rke2`):

| Datasource        | type                            | uid               | tracesToLogs today |
| ----------------- | ------------------------------- | ----------------- | ------------------ |
| VictoriaTraces    | tempo                           | `VictoriaTraces`  | **absent**         |
| VictoriaLogs (DS) | victoriametrics-logs-datasource | `VictoriaLogs`    | n/a                |
| VictoriaMetrics   | prometheus                      | `VictoriaMetrics` | n/a                |

Extend the VictoriaTraces entry in `infra/charts/observability` `defaultDatasources.extra` with `jsonData.tracesToLogsV2`:

```yaml
- name: VictoriaTraces
  type: tempo
  uid: VictoriaTraces
  url: http://vtsingle-vmks.monitoring.svc.cluster.local:10428/select/tempo
  jsonData:
    tracesToLogsV2:
      datasourceUid: VictoriaLogs
      spanStartTimeShift: "-5m"
      spanEndTimeShift: "5m"
      filterByTraceID: true
      filterBySpanID: false
      # If auto filterByTraceID is insufficient for VL LogSQL, enable customQuery
      # with a VL query using $${__trace.traceId} (escape $$ under Helm/Grafana).
```

Keep local `values.yaml` and `values-staging.yaml` aligned. Document LogSQL examples in `docs/observability.md` (exact VL filter syntax verified during apply against Explore).

**Alternatives considered:** custom Grafana app plugin; docs-only without tracesToLogs UI (worse DX).

### Decision: Docs live in `docs/observability.md`

Add a Structured logging section: field table, Trace → logs steps, redaction notes. Do not invent a separate ADR unless decisions expand later.

## Risks / Trade-offs

- **[Risk] VLAgent / Grafana JSON field nesting** (slog may nest attrs differently than flat LogSQL expects) → Mitigation: verify one checkout TraceId in Explore during apply; adjust handler options (e.g. replaceAttr) if needed.
- **[Risk] Log volume from access logs** → Mitigation: one line per HTTP/gRPC request; avoid Info-per-Kafka-success unless needed.
- **[Risk] Accidental sensitive payload logging** during migration → Mitigation: redaction helpers + spot-check payment/callback paths in test plan; never log raw request bodies by default.
- **[Trade-off] Fatal paths** use slog + exit instead of `log.Fatal` — slightly more verbose, consistent JSON.

## Migration Plan

1. Land `shared/observe/log` + `go.work` / CI globs / path filters.
2. Wire runtime bootstrap; migrate shared runtime + messaging; then service mains/kafka.
3. Web middleware swap; gRPC interceptor.
4. Observability chart tracesToLogs + docs.
5. Verify with `tilt up` traffic → VL Explore → Trace → logs.
6. Rollback: revert deploy; logs fall back to prior text only if image rolled back (no dual-write period).

## Open Questions

- Exact VictoriaLogs **LogSQL** expression for `customQuery` if `filterByTraceID: true` alone is insufficient with `victoriametrics-logs-datasource` — verify in Explore during apply (UID mapping is known: `VictoriaLogs`).
- Whether to map span tag `service.name` → log field `service` via `tracesToLogsV2.tags` once app logs consistently emit `service`.
