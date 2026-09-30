## Context

See proposal.md for motivation. The current state that shapes the approach:

- `talos-proxmox` runs an `otel-agent` DaemonSet in `observability` on both clusters. Its Service exposes OTLP 4317/4318 with `internalTrafficPolicy: Local`, so a pod reaches its own node's agent. The agent runs `k8sattributes` (with `otel_annotations: true`) and upserts `k8s.cluster.name` and `deployment.environment`, then forwards to prod's gateway, which writes the standard `otel_logs`, `otel_traces`, and `otel_metrics_*` tables that HyperDX reads.
- The agent's `logsCollection` preset reads `/var/log/pods/*` with only the `container` operator. JSON bodies stay unparsed strings, so `TraceId`, `SpanId`, and `SeverityText` are empty for marketplace logs.
- The agent's cluster collector scrapes pods annotated `prometheus.io/scrape` under the job name `annotated-pods`. The Prometheus receiver would use that as `service.name`, and the cluster collector does not run `k8sattributes`.
- There are no CiliumNetworkPolicies in `observability`. Marketplace policies are ingress-only default-deny, so egress to the agent is already open.
- `shared/observe/metric` builds a Prometheus-exporter MeterProvider served by an HTTP listener in `shared/runtime/metric.go`. All eight services call `runtime.InitMetrics`. `shared/observe/trace` hand-parses the endpoint and still carries VictoriaTraces `/insert/opentelemetry/` handling.
- The chart wires `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`, and `OTEL_TRACES_SAMPLER_ARG` from `defaults.otel.endpoint`, plus `METRICS_ADDR` and a `metrics` port from per-service `metrics`/`metricsPort` keys.
- Marketplace is not currently deployed on dev. The next root apply would fail on `VMPodScrape` and the `monitoring` namespace.

## Goals / Non-Goals

**Goals:**

- One OTLP path for traces and metrics, configured only through the standard `OTEL_*` environment variables.
- The same `service.name` on all three signals, so HyperDX filters and correlation work per service.
- Less code: no metrics listener, scrape resources, scrape policy, or dashboard JSON.

**Non-Goals:**

- Tuning sampling, adding custom business metrics, or changing span naming.
- Protocol choice per environment. Both clusters use OTLP/gRPC to the node agent.

## Decisions

### Endpoint: `http://otel-agent.observability.svc.cluster.local:4317`, OTLP/gRPC

```
 marketplace pod ──OTLP/gRPC──▶ otel-agent (same node) ──▶ prod otel-gateway ──▶ ClickHouse ◀── HyperDX
      │ stdout JSON                   ▲
      └──────── /var/log/pods ────────┘  (filelog + JSON parsing, talos-proxmox)
```

`defaults.otel.endpoint` becomes `http://otel-agent.observability.svc.cluster.local:4317`. The `http://` scheme is how the OTLP specification marks an insecure gRPC endpoint, and the Kafka Connect Java agent already uses that form. One value then serves the Go SDK and Connect. The chart keeps `OTEL_SERVICE_NAME` and `OTEL_TRACES_SAMPLER_ARG`, and adds `OTEL_METRIC_EXPORT_INTERVAL=30000`.

Alternative considered: OTLP/HTTP on 4318. It works equally well, but gRPC is what the exporters and Connect already use.

### Go exporters read the standard environment variables

`shared/observe/trace` keeps its no-op path when `OTEL_EXPORTER_OTLP_ENDPOINT` is empty, which tests and host processes rely on. Otherwise it calls `otlptracegrpc.New(ctx)` without endpoint options, so the SDK applies `OTEL_EXPORTER_OTLP_*` itself, including the insecure `http://` scheme. `UseHTTP`, the `/insert/opentelemetry/` rewriting, `defaultOTLPEndpoint`, and `DefaultEndpoint()` are deleted.

`shared/observe/metric` replaces the Prometheus exporter with `sdkmetric.NewPeriodicReader(otlpmetricgrpc.New(ctx))`, using the same empty-endpoint rule. With no endpoint it installs a MeterProvider without a reader, so instruments stay valid and export nothing. The resource is `resource.Default()` merged with `service.name`, matching tracing. `Config.Addr`, `METRICS_ADDR`, `Handler()`, and the `prometheus`/`client_golang` dependencies are removed.

`runtime.InitMetrics` keeps its signature, so service `main` files do not change. It no longer starts a listener, and it returns the MeterProvider's shutdown, which flushes pending metrics.

The default cumulative temporality is kept. HyperDX computes rates and quantiles from cumulative sums and histograms in the exporter's schema.

Alternative considered: `go.opentelemetry.io/contrib/exporters/autoexport`. It selects exporters from `OTEL_*_EXPORTER`, but falls back to `localhost:4318` when the endpoint is unset, which conflicts with the no-op default, and adds a dependency for one supported protocol.

Alternative considered: keep `/metrics` and use the platform's annotation scrape. Every series would get `service.name=annotated-pods` and no pod metadata, and a scrape policy for `observability` would have to be kept.

### Service identity through the `resource.opentelemetry.io/service.name` annotation

`services.tpl` adds `resource.opentelemetry.io/service.name: <name>` to every Deployment's pod template, and the kafka chart adds `connect-debezium` to the KafkaConnect pod template. The agent's `k8sattributes` reads it through `otel_annotations`, so filelog records get the right `ServiceName`. For pushed signals it carries the same value the SDK already sets. The `app: <name>` label is kept for selectors and policies.

Alternative considered: rely on `k8sattributes`' fallback from workload names. It depends on processor-version defaults and silently varies between Deployments and StatefulSets.

### JSON log parsing lives in `talos-proxmox`

The in-flight `talos-proxmox` change `add-clickstack-observability` adds operators after `container-parser` in the `otel-agent` filelog receiver, each one conditional on its key being present and set to `on_error: send_quiet`:

1. `json_parser` only when the body starts with `{`, parsing into `attributes`.
2. Severity from `attributes.level` (slog's `DEBUG`/`INFO`/`WARN`/`ERROR`).
3. `trace_parser` from `attributes.trace_id`/`attributes.span_id`.
4. Move `attributes.msg` to `body`.

The container runtime timestamp is kept instead of parsing `time`, because the two differ by microseconds and parsing adds a failure mode for other workloads' JSON.

Non-JSON lines, and JSON without these keys, pass through unchanged. The marketplace contract is the key set in the structured-logging spec, which slog already emits, so `shared/observe/log` does not change.

Alternative considered: the `otelslog` bridge exporting logs over OTLP. Every line would be stored twice, once over OTLP and once from stdout, unless the platform excluded `ecommerce` from filelog, which couples the platform to this namespace.

### Chart cleanup

Delete `podscrape.tpl`, `dashboards.tpl`, `dashboards/`, the `allow-metrics-scrape` block in `mesh-policy.tpl`, `metricsScrape`, the `metrics`/`metricsPort` service keys (including the simulator's `metrics: false`), the `marketplace.metrics` label, the `metrics` container and Service ports, and `METRICS_ADDR`.

## Risks / Trade-offs

- [Marketplace logs lack `TraceId` until the talos-proxmox parsing lands] → Marketplace can merge first. Logs are still searchable by body text and service, and trace-to-log verification waits for the platform change.
- [Cluster-wide JSON parsing affects other workloads' logs] → Parse only bodies starting with `{`, `on_error: send_quiet`, and promote only when keys exist, so other formats keep today's behavior.
- [HyperDX metric views do not render OTel histogram names or cumulative temporality as expected] → Verify p95 of `http.server.request.duration` on prod before removing the old docs. If that fails, switch the reader to delta temporality, which is a one-line change.
- [The `service.name` annotation depends on the platform keeping `otel_annotations`] → Document it as part of the platform contract in `docs/deployment/observability.md`.
- [Node agent unavailable during a rollout] → The SDK batch processor and periodic reader drop data after retries. Services keep serving.
- [No Git-managed dashboards] → Accepted. The docs carry the RED and log queries.

## Migration Plan

1. Land the `talos-proxmox` JSON parsing change. It is independent and can go first or in parallel.
2. Merge this change and apply the prod marketplace root, which syncs with no Victoria resources on `:main`. Marketplace stays undeployed on dev.
3. Verify on prod in HyperDX: spans for all services and `connect-debezium`, RED metrics per service, logs with `ServiceName`, and, once step 1 is live, trace-to-log navigation.
4. When archiving, update the `## Purpose` sections of `platform-observability`, `app-otel-metrics`, `distributed-tracing`, and `cilium-observability` to drop the Victoria and Grafana wording.

Rollback: revert the merge. The Victoria-based chart cannot sync on the current platform, so rollback only restores the previous code, not working telemetry.
