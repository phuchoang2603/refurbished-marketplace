# Marketplace observability

The shared telemetry pipeline is installed and administered by `talos-proxmox`: prod's VictoriaMetrics, VictoriaLogs, VictoriaTraces, VMAgent, VLAgent, VTAgent and Grafana live in [`apps/components/observability`](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps/components/observability). Dev runs no telemetry components. This repository owns only the marketplace telemetry producers.

## Application contract

| Signal  | Marketplace output                                                                                                     | Platform consumer         |
| ------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------- |
| Traces  | OTLP/HTTP to `http://vtagent-vmks.observability.svc.cluster.local:10429/insert/opentelemetry/v1/traces`                | VTAgent → VictoriaTraces  |
| Metrics | Cumulative OTLP/HTTP to `http://vmagent-vmks.observability.svc.cluster.local:8429/opentelemetry/v1/metrics` every 30 s | VMAgent → VictoriaMetrics |
| Logs    | Single-line slog JSON on stdout                                                                                        | VLAgent → VictoriaLogs    |

Charts set the per-signal `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` and `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`, plus `k8s.cluster.name=prod-talos` and `deployment.environment=prod` resource attributes. Applications do not depend on an OpenTelemetry collector. VMAgent and VTAgent persist bounded remote-write queues on a separate node from the stores and replay after storage outages. VLAgent attaches Kubernetes metadata to container logs.

## Platform prerequisites

Prod's `observability` Application (agents, collectors, ClickHouse, MongoDB and HyperDX) must be healthy for data to be stored. Its critical namespace scope must include `ecommerce`; a healthy exporter alone does not prove the platform retains its data:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get applications -n argo-cd observability
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get pods,svc -n observability
```

Fetch kubeconfigs as described in [gitops.md](gitops.md#environment-configuration). Retention (7 days), storage, ingest authentication, and HyperDX accounts are documented in `talos-proxmox`'s [cluster access](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/operations/cluster-access.md) guide.

## HyperDX access

HyperDX is served at <https://hyperdx.phuchoang.sbs>. The public tunnel and LAN service at `http://10.69.12.128` both use the operator-only Dex access proxy. HyperDX then requires its own account login; see the platform [access guide](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/operations/cluster-access.md#public-urls-prod).

The default sources are **Logs** (`otel_logs`), **Traces** (`otel_traces`), and **Metrics** (`otel_metrics_gauge`, `otel_metrics_sum`, `otel_metrics_histogram`).

## Verify application telemetry

Confirm the exporter configuration on the pods:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get deploy -n ecommerce \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[-1:].env[?(@.name=="OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")].value}{"\n"}{end}'
```

Each service logs `tracing enabled` and `metrics export enabled` with the endpoint at startup.

In HyperDX, open **Chart Explorer** or the SQL editor and check that each service sends every signal within the last 15 minutes:

Use Grafana Explore with VictoriaMetrics, VictoriaLogs, and VictoriaTraces to inspect recent service telemetry. Exercise the shop before checking RED metrics.

## Application RED metrics

HTTP (web) and gRPC RED come from OpenTelemetry instrumentation:

| Metric                         | Unit | Useful attributes                                                |
| ------------------------------ | ---- | ---------------------------------------------------------------- |
| `http.server.request.duration` | s    | `http.route`, `http.request.method`, `http.response.status_code` |
| `rpc.server.call.duration`     | s    | `rpc.method`, `rpc.response.status_code`                         |
| `rpc.client.call.duration`     | s    | `rpc.method`, `rpc.response.status_code`                         |

In Grafana Explore, select VictoriaMetrics and use PromQL to query request rates, errors, latency histograms, and the Checkout saga gauges. Metric names are normalized to Prometheus format. A zero gauge confirms it is queryable, not that failure paths were exercised. To verify accepted workflows, read Checkout by buyer-authenticated `GetCheckout` and cross-check persisted `checkouts`/`checkout_exceptions`.

## Distributed tracing

Application tracing is initialized through `shared/observe/trace` and `shared/runtime`. Services propagate W3C `traceparent`; outbox rows carry `tracingspancontext`, Debezium maps it to Kafka headers, and consumers continue the trace.

```text
Browser → web → domain services → database
                                                              │
                    └→ outbox → Debezium/Connect → Kafka → consumers
                                                              │
                                  VTAgent ◀──────────────────┘ → VictoriaTraces → Grafana
```

KafkaConnect enables Strimzi OpenTelemetry tracing and exports OTLP/HTTP to the same VTAgent traces endpoint. Rebuild the Connect image only when its plugin contents change.

To verify a checkout, search Grafana Explore with the VictoriaTraces data source:

```text
service.name = checkout
```

Open a `POST /checkout` span. Expect route/RPC/messaging span names, database child spans where applicable, and `connect-debezium` across asynchronous hops. Gateway proxy spans are not expected.

For payment outcomes, follow `POST /webhooks/payment-gateway` → Checkout `HandleGatewayWebhook` → `checkout_outbox` → Debezium → Inventory commit/release. Orders, Inventory and Checkout keep their tracing outbox columns; there is no Payment outbox. Correlate the gateway delivery with `order_id` and gateway payment ID. Never log the API key, webhook signing secret, hosted checkout URL/token, or raw webhook body. Cluster verification of this waterfall remains pending; see the [rollout gate](../order-placement.md#rollout-gate).

## Structured logging and correlation

Marketplace services emit JSON slog lines via `shared/observe/log`. Request-path logging should use context-aware helpers so `trace_id` and `span_id` are present. Sensitive keys are redacted, but free-text messages are not rewritten; never put secrets or payment payloads in log messages.

Common fields include `service`, `trace_id`, `span_id`, HTTP/gRPC method and status, Kafka topic/partition/offset, and domain IDs such as `order_id`.

VictoriaLogs LogsQL search examples:

```text
service:=orders trace_id:=<hex-trace-id>
```

```text
order_id:"<order-id>"
```

```text
kubernetes.pod_namespace:=ecommerce level:=ERROR
```

For a failed checkout, open the trace first; the trace side panel lists log records that share its `trace_id`. From a log record, use its trace link to open the waterfall, then narrow by `order_id` if needed. Align the time range with the trace before concluding that logs are missing.
