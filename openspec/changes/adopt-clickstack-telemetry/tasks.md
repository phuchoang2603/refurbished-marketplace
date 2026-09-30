## 1. Platform JSON log parsing (talos-proxmox)

- [x] 1.1 In `../talos-proxmox`, extend the in-flight `add-clickstack-observability` change (design bullet and task 4.7) with the design's filelog operators to `apps/components/otel-agent/values.yaml`: `json_parser` gated on a body starting with `{` with `on_error: send_quiet`, severity from `level`, `trace_parser` from `trace_id`/`span_id`, and moving `msg` to the body; verify `helm template` renders the operators after `container-parser`
- [ ] 1.2 Land that change and verify on prod that a JSON line from any pod is stored with a populated `SeverityText`, a `Body` equal to its `msg`, and a `TraceId` when the line had `trace_id`

## 2. Go tracing bootstrap

- [x] 2.1 In `shared/observe/trace`, keep the no-op path for an empty `OTEL_EXPORTER_OTLP_ENDPOINT`, build the exporter with `otlptracegrpc.New(ctx)` so it reads `OTEL_EXPORTER_OTLP_*`, and delete `UseHTTP`, the `/insert/opentelemetry/` rewriting, `defaultOTLPEndpoint`, `DefaultEndpoint()`, and the unused `otlptracehttp` import; verify `rg -n "insert/opentelemetry|vtsingle|DefaultEndpoint" shared services` returns nothing and `go build ./...` passes in `shared/observe/trace`
- [x] 2.2 Run `devenv tasks run go:tidy`; verify `go.mod` in `shared/observe/trace` no longer requires `otlptracehttp` and the workspace builds

## 3. Go metrics bootstrap

- [x] 3.1 In `shared/observe/metric`, replace the Prometheus exporter with a periodic reader over `otlpmetricgrpc.New(ctx)`, install a reader-less MeterProvider when `OTEL_EXPORTER_OTLP_ENDPOINT` is empty, and remove `Config.Addr`, `METRICS_ADDR`, `defaultMetricsAddr`, and `Handler()`; verify `go build ./...` passes in the module
- [x] 3.2 In `shared/runtime/metric.go`, drop the HTTP listener and return the MeterProvider shutdown, logging whether metric export is enabled; verify `rg -n "METRICS_ADDR|/metrics|promhttp" shared services` returns nothing
- [x] 3.3 Run `devenv tasks run go:tidy`; verify no module in the workspace requires `go.opentelemetry.io/otel/exporters/prometheus` or `github.com/prometheus/client_golang` directly, and `go build ./...` passes for every service module
- [x] 3.4 Run the existing tests for `shared/observe/...`, `shared/runtime`, and `services/web`; verify they pass

## 4. Marketplace chart

- [x] 4.1 In `infra/charts/refurbished-marketplace/values.yaml`, set `defaults.otel.endpoint` to `http://otel-agent.observability.svc.cluster.local:4317` and remove `metricsScrape` and every service `metrics`/`metricsPort` key; verify `rg -n "metricsScrape|metricsPort|metrics:" infra/charts/refurbished-marketplace/values*.yaml` returns nothing
- [x] 4.2 In `templates/services.tpl`, remove the `marketplace.metrics` label, the `metrics` container and Service ports, and `METRICS_ADDR`; add `OTEL_METRIC_EXPORT_INTERVAL: "30000"` and the pod annotation `resource.opentelemetry.io/service.name: <name>`; verify `helm template` shows each Deployment with the annotation, the new endpoint, and no port 9100
- [x] 4.3 Delete `templates/podscrape.tpl`, `templates/dashboards.tpl`, and `dashboards/`, and remove the `allow-metrics-scrape` block from `templates/mesh-policy.tpl`; verify `helm template` output contains no `VMPodScrape`, `monitoring`, `grafana_dashboard`, or `allow-metrics-scrape`
- [x] 4.4 Render with `values-prod.yaml` as well; verify both renders succeed and `helm lint` passes

## 5. Kafka chart

- [x] 5.1 In `infra/charts/kafka/templates/connect.tpl`, change the Connect OTLP default to `http://otel-agent.observability.svc.cluster.local:4317` and add the pod annotation `resource.opentelemetry.io/service.name: connect-debezium` through the KafkaConnect pod template; verify `helm template` of the kafka chart shows both and no `vtsingle`

## 6. Documentation

- [x] 6.1 Rewrite `docs/deployment/observability.md` for HyperDX: the telemetry contract table (OTLP to `otel-agent`, stdout JSON, `service.name` annotation), the internal HyperDX address, per-signal arrival checks, HyperDX searches by service, trace ID, and `order_id`, trace-to-log steps, RED chart queries on `http.server.request.duration` and the RPC duration metric, and the promoted log keys; verify `rg -in "victoria|grafana|vmagent|logsql|traceql|9100" docs/deployment/observability.md` returns nothing
- [x] 6.2 Update `docs/deployment/gitops.md` (component table, platform ownership line, CRD check), `docs/deployment/cilium.md` (observe-path paragraph and Grafana Gateway/tunnel rows), `docs/architecture.md` (Observe row), and `docs/development/code-generation.md` (the `shared/observe/metric` row); verify `rg -in "victoria|vmpodscrape|grafana|monitoring|:9100" docs --glob '*.md'` returns nothing
- [x] 6.3 Update `docs/diagrams/architecture.excalidraw` and the matching text in `architecture.svg` so the observe path shows `otel-agent` → prod ClickStack; verify neither file mentions Victoria

## 7. Rollout verification (prod)

- [ ] 7.1 Apply the prod marketplace root on the fresh prod cluster; verify every marketplace Application is Synced and Healthy, and no pod exposes port 9100
- [ ] 7.2 Place a checkout on `shop.phuchoang.sbs`; verify HyperDX shows one TraceId across `web`, `orders`, `connect-debezium`, `inventory`, and `payment`, with `deployment.environment=prod`
- [ ] 7.3 Verify HyperDX metrics show `http.server.request.duration` for `web` and RPC duration for the gRPC services, filtered by `ServiceName`, and that a p95 chart renders; if histograms do not render, switch the reader to delta temporality and re-verify
- [ ] 7.4 Verify marketplace logs in HyperDX have the correct `ServiceName` and, with task 1.2 done, that opening the checkout trace lists its log lines

## 8. Archive preparation

- [ ] 8.1 Update the `## Purpose` sections of `openspec/specs/platform-observability`, `app-otel-metrics`, `distributed-tracing`, and `cilium-observability` to describe the OTLP, `otel-agent`, and HyperDX path; verify `rg -in "victoria|grafana" openspec/specs` returns nothing after archive
