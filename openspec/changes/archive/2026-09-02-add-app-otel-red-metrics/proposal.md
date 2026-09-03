## Why

Issue [#43](https://github.com/phuchoang2603/refurbished-marketplace/issues/43): after Istio removal, there is no request/error/duration SLI path. Traces still go app OTEL → VictoriaTraces; HTTP/gRPC instrumentation already records RED against a noop meter. Operators need application-layer RED in VictoriaMetrics and Grafana without Hubble or Istio scrapes.

## What Changes

- Install a shared OpenTelemetry **MeterProvider** next to the existing tracer bootstrap so `otelhttp` (web) and `otelgrpc` (gRPC servers and clients) export HTTP/gRPC RED to VictoriaMetrics over OTLP/HTTP.
- Split export destinations: traces stay on VTSingle (`:4317`); metrics go to VMSingle (`/opentelemetry/v1/metrics`). Do not send metrics to VictoriaTraces or introduce an OpenTelemetry Collector.
- Wire marketplace Helm with a dedicated metrics OTLP endpoint (keep the existing traces env). No per-service Prometheus `/metrics` ports or VMPodScrapes for apps.
- Enable repo-owned Grafana dashboards and add a Marketplace RED board (rate, error ratio, latency by service and route/RPC).
- Update `docs/observability.md` (and related observe notes) so app OTEL metrics are the RED path; stop pointing operators at #43 or Hubble.

Non-goals: Hubble (relay/UI/metrics) and Cilium L7 visibility; OpenTelemetry Collector / spanmetrics / VMAgent OTLP fan-out; per-pod `/metrics`; payment-gateway-simulator instrumentation; Kafka consumer RED; canary/version split metrics; Gateway/Envoy edge SLIs; restoring Marketplace Istio RED or `hubble_http_*` dashboards.

## Capabilities

### New Capabilities

- `app-otel-metrics`: Marketplace services export HTTP/gRPC RED via OTEL metrics to VictoriaMetrics; Grafana shows a Marketplace RED dashboard; Hubble and Istio are not the source.

### Modified Capabilities

- `cilium-observability`: App RED is no longer deferred to #43. Hubble remains not required; Istio scrapes/RED stay absent. Documentation describes app OTEL metrics instead of a follow-on issue.
- `platform-observability`: Marketplace workloads SHALL export OTLP metrics to VMSingle when enabled. Custom `/metrics` endpoints remain out of scope. Grafana SHALL load a marketplace RED dashboard. Docs cover metrics Explore / dashboard use.

## Impact

- Go: `shared/observe` (new metrics bootstrap), `shared/runtime` init/shutdown, existing `otelhttp`/`otelgrpc` wiring (no duplicate histograms unless route cardinality is wrong).
- Helm: marketplace OTEL env; observability chart `customDashboards` + dashboard JSON.
- Docs: `docs/observability.md`; drop #43-as-gap language in Cilium observe notes where this change lands.
- Does not change Cilium Helm (talos-proxmox), CNPs, Kafka Connect, protobuf, or business logic.
