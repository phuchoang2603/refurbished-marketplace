## Context

See proposal.md for motivation. Marketplace services already call `runtime.InitTracing` and use `otelhttp` / `otelgrpc`. Those libraries record RED on the **global MeterProvider**, which is still the SDK noop. Traces use `OTEL_EXPORTER_OTLP_ENDPOINT` → `vtsingle-vmks.monitoring.svc.cluster.local:4317`. VMSingle already stores cluster metrics; it accepts OTLP metrics at HTTP `POST /opentelemetry/v1/metrics` (port `8428` on the VMSingle Service). `customDashboards` on the observability chart is off and `dashboards/` is empty. Specs: `app-otel-metrics`, plus deltas on `cilium-observability` and `platform-observability`.

## Goals / Non-Goals

**Goals:**

- One MeterProvider bootstrap reused by all marketplace Go services that already init tracing (`web`, `users`, `products`, `orders`, `cart`, `payment`).
- Keep trace and metric OTLP destinations independent in env and Helm.
- Reuse existing HTTP/gRPC instrumentation for series; Grafana PromQL against VictoriaMetrics after OTEL name/label mapping.
- Confirm HTTP metrics use chi route patterns (`/orders/{id}`), not raw paths.

**Non-Goals:**

- Collector, spanmetrics, VMAgent OTLP receive, Hubble, canary labels, Kafka/simulator RED, `/metrics` scrape (see proposal non-goals).
- Changing trace sampling, span names, or the VT ingest path.

## Decisions

### 1. New `shared/observe/metric` plus `runtime.InitMetrics`

Add a metrics package next to `shared/observe/trace` (resource `service.name`, OTLP/HTTP exporter, periodic reader, `otel.SetMeterProvider`). `runtime.InitMetrics` mirrors `InitTracing`: load env, init, log enabled/disabled, return shutdown. Each `main` that already inits tracing also inits metrics and defers shutdown.

**Rationale:** Trace bootstrap stays traces-only; mixing exporters in `trace.go` would send metrics to VT if someone reused `OTEL_EXPORTER_OTLP_ENDPOINT`. Separate files match repo conventions.

**Alternatives considered:** OpenTelemetry Collector fan-out (extra pod, same Go MeterProvider still required); spanmetrics (no MeterProvider, but cardinality/sampling and a new collector); Prometheus `/metrics` + scrape (new ports, contradicts “no scrape for app RED”).

### 2. Metrics endpoint is OTLP/HTTP to VMSingle, not VT `:4317`

Default in-cluster URL: `http://vmsingle-vmks.monitoring.svc.cluster.local:8428/opentelemetry/v1/metrics` (confirm Service name/port on the live chart if it differs). Helm: `defaults.otel.metricsEndpoint` (or equivalent) rendered as `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`. Do **not** set a shared `OTEL_EXPORTER_OTLP_ENDPOINT` for metrics; leave traces env as today. Empty metrics endpoint → noop MeterProvider (process starts).

Use `otlpmetrichttp` with the full URL. Cumulative temporality (Go SDK default) matches VictoriaMetrics.

**Rationale:** VT OTLP gRPC is traces-only. VM documents protobuf OTLP metrics on that HTTP path.

**Alternatives considered:** Push to VMAgent OTLP (extra hop, same app code); one collector URL (rejected for this change).

### 3. Do not add a second HTTP/gRPC histogram

Keep `otelhttp.NewMiddleware` and `otelgrpc` stats handlers. After MeterProvider is global, they export `http.server.request.duration` / `rpc.server.duration` / `rpc.client.duration` (semconv names as implemented by current contrib). If HTTP metrics lack `http.route` (chi pattern), fix the existing middleware (metric attributes from `r.Pattern` / route context)—do not record a parallel custom histogram.

**Rationale:** Duplicate series confuse the dashboard and double cardinality.

**Alternatives considered:** Manual chi interceptor (more code, diverges from traces); log-derived RED (not histograms).

### 4. Grafana: enable `customDashboards` and one Marketplace RED JSON

Turn `customDashboards.enabled` on. Add `infra/charts/observability/dashboards/*.json` for HTTP (web) and gRPC (by `service` / RPC). Query VictoriaMetrics; filter marketplace services (`web`, `users`, `products`, `orders`, `cart`, `payment`). Error ratio from HTTP 5xx / gRPC non-OK. Latency as histogram quantile. Build PromQL against **actual** series after a first export (VM may underscore OTEL names and copy resource attrs to labels). Do not query `hubble_http_*` or `istio_requests_total`.

**Rationale:** Chart already glob-loads dashboard ConfigMaps; sidecar + `grafana_folder: Marketplace` already exist.

**Alternatives considered:** Grafana Explore-only (no durable SLI view); import Istio RED JSON (wrong metrics).

### 5. Docs as the operator contract

Update `docs/observability.md` (and Cilium observe sentences that still cite #43 as the gap) with the two OTLP destinations, dashboard name, and “Hubble off.” Keep traces/logs sections.

**Rationale:** Specs require documentation of the RED path.

## Risks / Trade-offs

- **[Metrics pointed at VT :4317]** → Separate env var; never default metrics to `defaults.otel.endpoint`.
- **[HTTP cardinality / missing route]** → Verify `http.route` on a checkout path; fix otelhttp attributes if the series uses `/orders/<uuid>`.
- **[PromQL vs OTEL names]** → Draft dashboard after one live ingest; adjust label names (`service_name` vs `job`) to match VM mapping.
- **[OTLP push loss if VMSingle down]** → Periodic export + SDK retry; same class of failure as traces to VT. No scrape fallback in this change.
- **[Client + server gRPC double-count]** → Dashboard uses **server** RPC for service SLIs; optional client panel for `web` outbound only, labeled as such.

## Migration Plan

1. Merge Go + Helm + dashboard + docs; Argo syncs observability then marketplace (or same wave).
2. Old images without MeterProvider stay traces-only until rolled; Grafana panels empty until new pods export.
3. Rollback: unset metrics endpoint or revert images; traces/logs unchanged. Dashboard ConfigMap can stay (empty queries).
4. No Hubble or Cilium Helm change. Close [#43](https://github.com/phuchoang2603/refurbished-marketplace/issues/43) when dashboard shows live checkout/gRPC traffic on talos-dev.
