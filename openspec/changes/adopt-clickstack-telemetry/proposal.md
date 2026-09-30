## Why

`talos-proxmox` removed VictoriaMetrics, VictoriaLogs, VictoriaTraces, and Grafana (`remove-longhorn-and-victoria`) and replaced them with ClickStack (`add-clickstack-observability`): an OpenTelemetry agent in each cluster at `otel-agent.observability.svc`, forwarding to one ClickHouse store with the HyperDX UI on prod. The marketplace chart still renders a `VMPodScrape` (CRD gone) and dashboard ConfigMaps in `monitoring` (namespace gone), so its next sync fails, and its traces target a Service that no longer exists.

## What Changes

- Export traces from Go services and Kafka Connect over OTLP to the node-local `otel-agent` instead of `vtsingle-vmks`.
- **BREAKING** Replace the Prometheus `/metrics` scrape with OTLP metric push to the same agent. Remove the `:9100` listener, `METRICS_ADDR`, the `metrics` container and Service ports, the `marketplace.metrics` pod label, the `metrics`/`metricsPort` chart keys, and the `metricsScrape` value.
- **BREAKING** Remove `VMPodScrape/marketplace-apps`, the `allow-metrics-scrape` CiliumNetworkPolicy, and the Grafana dashboard ConfigMaps (Marketplace RED and Marketplace logs) together with their JSON sources.
- Remove the VictoriaTraces-specific `/insert/opentelemetry/` HTTP path handling and the `vtsingle` default from `shared/observe/trace`.
- Keep JSON logs on stdout. Rely on a `talos-proxmox` change that makes `otel-agent` parse JSON log bodies and promote `trace_id`, `span_id`, `level`, and `msg` into the OTLP trace, span, severity, and body fields, so HyperDX can link traces to logs.
- Rewrite observability documentation for HyperDX: access, per-signal verification, trace-to-log navigation, and example searches that replace the removed dashboards.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `platform-observability`: the platform contract becomes OTLP to `otel-agent` for traces and metrics plus stdout JSON for logs, viewed in HyperDX. Victoria, VMAgent, and Grafana dashboard requirements are removed.
- `app-otel-metrics`: RED metrics are pushed over OTLP instead of scraped from `/metrics`; the Grafana RED dashboard requirement is removed.
- `distributed-tracing`: the export destination moves from VictoriaTraces to the platform OTLP agent, and verification moves to HyperDX.
- `structured-logging`: documentation moves from VictoriaLogs LogsQL to HyperDX search, and the log contract states which JSON keys the platform promotes.
- `cilium-observability`: the metrics scrape policy exception is removed, and traces and RED references move off Victoria.
- `web`: web traces export to the platform OTLP agent instead of VictoriaTraces.
- `cilium-ingress`: the Cloudflare Tunnel origin requirement drops its leftover Grafana Service reference.

## Non-goals

- Provisioning HyperDX dashboards, saved searches, or alerts. HyperDX stores them in its own MongoDB, outside Git.
- OTLP log export from the application (`otelslog` bridge). stdout stays the only log path.
- Metrics from Kafka, CloudNativePG, MongoDB, Meilisearch, or Redis.
- Any change to the ClickHouse store, the gateway, or ingest authentication in `talos-proxmox` beyond the agent's JSON log parsing.

## Impact

- **Go:** `shared/observe/metric` (Prometheus exporter replaced by an OTLP periodic reader), `shared/observe/trace` (legacy Victoria handling removed), and `shared/runtime/metric.go` (no HTTP listener). Service `main` files keep calling `runtime.InitMetrics`.
- **Chart `refurbished-marketplace`:** `values.yaml`, `services.tpl`, and `mesh-policy.tpl` change. `podscrape.tpl`, `dashboards.tpl`, and `dashboards/*.json` are removed.
- **Chart `kafka`:** the Connect OTLP endpoint changes in `connect.tpl`.
- **Cross-repo:** the `talos-proxmox` `otel-agent` component gains a JSON log parser. It must be merged before log correlation is verified, but marketplace can sync without it.
- **Docs:** `docs/deployment/observability.md`, `docs/deployment/gitops.md`, `docs/deployment/cilium.md`, `docs/architecture.md`, and the architecture diagram.
