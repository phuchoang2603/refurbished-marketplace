# Marketplace observability

The shared telemetry pipeline is installed and administered by `talos-proxmox`: prod's OpenTelemetry node agents, cluster collector, gateway and ClickHouse store with the HyperDX UI live in [`apps/components/observability`](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps/components/observability). Dev runs no telemetry components. This repository owns only the marketplace telemetry producers.

## Application contract

| Signal  | Marketplace output                                                        | Platform consumer                              |
| ------- | ------------------------------------------------------------------------- | ---------------------------------------------- |
| Traces  | OTLP/gRPC to `http://otel-agent.observability.svc.cluster.local:4317`     | `otel-agent` → prod gateway → `otel_traces`    |
| Metrics | OTLP/gRPC to the same endpoint every 30 s (`OTEL_METRIC_EXPORT_INTERVAL`) | `otel-agent` → prod gateway → `otel_metrics_*` |
| Logs    | single-line slog JSON on stdout                                           | `otel-agent` filelog → `otel_logs`             |

The `otel-agent` Service uses `internalTrafficPolicy: Local`, so each pod exports to its own node's agent without credentials. The agent adds Kubernetes metadata and sets `k8s.cluster.name` and `deployment.environment` (`prod`); the chart does not set them.

Every marketplace pod template carries `resource.opentelemetry.io/service.name: <service>`. The agent's `k8sattributes` reads it, so stdout log records get the same `ServiceName` as the service's spans and metrics. KafkaConnect pods carry `connect-debezium`.

The agent parses log bodies that start with `{` and promotes these slog keys:

| slog key   | Stored as      |
| ---------- | -------------- |
| `level`    | `SeverityText` |
| `msg`      | `Body`         |
| `trace_id` | `TraceId`      |
| `span_id`  | `SpanId`       |

All other keys (`service`, `order_id`, …) stay in `LogAttributes`. Keep those five keys stable; renaming them breaks trace-to-log navigation.

The marketplace charts render no scrape resources, dashboards, or policies for telemetry, and services open no metrics port. OTLP egress from ambient-enrolled pods to the non-enrolled `otel-agent` leaves through ztunnel as plaintext. Collecting Istio ingress, waypoint, and ztunnel retry, overflow, and outlier-ejection signals is the platform-owned agent's job in `talos-proxmox`; those signals stay separate from application spans and RED metrics.

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
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[-1:].env[?(@.name=="OTEL_EXPORTER_OTLP_ENDPOINT")].value}{"\n"}{end}'
```

Each service logs `tracing enabled` and `metrics export enabled` with the endpoint at startup.

In HyperDX, open **Chart Explorer** or the SQL editor and check that each service sends every signal within the last 15 minutes:

```sql
SELECT * FROM (
SELECT 'traces' AS signal, ServiceName, count() FROM otel_traces
WHERE Timestamp > now() - INTERVAL 15 MINUTE AND ResourceAttributes['k8s.namespace.name'] = 'ecommerce'
GROUP BY ServiceName
UNION ALL
SELECT 'logs', ServiceName, count() FROM otel_logs
WHERE Timestamp > now() - INTERVAL 15 MINUTE AND ResourceAttributes['k8s.namespace.name'] = 'ecommerce'
GROUP BY ServiceName
UNION ALL
SELECT 'metrics', ServiceName, count() FROM otel_metrics_histogram
WHERE TimeUnix > now() - INTERVAL 15 MINUTE AND ResourceAttributes['k8s.namespace.name'] = 'ecommerce'
GROUP BY ServiceName
) ORDER BY signal, ServiceName
```

Expect `web`, `users`, `products`, `inventory`, `orders`, `checkout`, `cart`, and `search` for all three signals. A service exports RED metrics only after it handles its first request, so exercise the shop before concluding metrics are missing. If a service is missing, check its startup logs for exporter errors and the `otel-agent` pod on the same node.

## Application RED metrics

HTTP (web) and gRPC RED come from OpenTelemetry instrumentation:

| Metric                         | Unit | Useful attributes                                                |
| ------------------------------ | ---- | ---------------------------------------------------------------- |
| `http.server.request.duration` | s    | `http.route`, `http.request.method`, `http.response.status_code` |
| `rpc.server.call.duration`     | s    | `rpc.method`, `rpc.response.status_code`                         |
| `rpc.client.call.duration`     | s    | `rpc.method`, `rpc.response.status_code`                         |

In HyperDX **Chart Explorer**, pick the **Metrics** source and the histogram metric, then:

- request rate: aggregation **Count** (per-second rate), grouped by `ServiceName` or `http.route`;
- error ratio: the same chart filtered to `Attributes['http.response.status_code'] >= '500'` (HTTP) or `Attributes['rpc.response.status_code'] != 'OK'` (gRPC), compared with the unfiltered count;
- latency: aggregation **p95**, grouped by `ServiceName`.

Use HyperDX **Chart Explorer** to inspect the existing `checkout.pending.max_age.seconds`, `checkout.compensating.count`, `checkout.financial_exceptions.count`, and `checkout.stuck.count` gauges, filtered to `ServiceName = checkout` and `deployment.environment = prod`. The platform supplies bundled preset dashboards and does not provision a custom **Prod Checkout Saga** dashboard. A zero reading confirms the gauge is queryable, not that failure paths were exercised. Dashboards live in HyperDX's own database, not in Git. To verify accepted workflows, read Checkout by buyer-authenticated `GetCheckout` and cross-check persisted `checkouts`/`checkout_exceptions`; never infer completion from a browser redirect.

## Distributed tracing

Application tracing is initialized through `shared/observe/trace` and `shared/runtime`. Services propagate W3C `traceparent`; outbox rows carry `tracingspancontext`, Debezium maps it to Kafka headers, and consumers continue the trace.

```text
Browser → web → domain services → database
                                                              │
                    └→ outbox → Debezium/Connect → Kafka → consumers
                                                              │
                                  node otel-agent ◀────────────┘ → prod ClickHouse → HyperDX
```

KafkaConnect enables Strimzi OpenTelemetry tracing and exports to the same `otel-agent` endpoint. Rebuild the Connect image only when its plugin contents change.

To verify a checkout, search the **Traces** source:

```text
ServiceName:(web OR checkout OR orders OR products OR inventory OR search OR cart OR users OR connect-debezium)
```

Open a `POST /checkout` span. Expect route/RPC/messaging span names, database child spans where applicable, and `connect-debezium` across asynchronous hops. Gateway proxy spans are not expected.

For payment outcomes, follow `POST /webhooks/payment-gateway` → Checkout `HandleGatewayWebhook` → `checkout_outbox` → Debezium → Inventory commit/release. Orders, Inventory and Checkout keep their tracing outbox columns; there is no Payment outbox. Correlate the gateway delivery with `order_id` and gateway payment ID. Never log the API key, webhook signing secret, hosted checkout URL/token, or raw webhook body. Cluster verification of this waterfall remains pending; see the [rollout gate](../order-placement.md#rollout-gate).

## Structured logging and correlation

Marketplace services emit JSON slog lines via `shared/observe/log`. Request-path logging should use context-aware helpers so `trace_id` and `span_id` are present. Sensitive keys are redacted, but free-text messages are not rewritten; never put secrets or payment payloads in log messages.

Common fields include `service`, `trace_id`, `span_id`, HTTP/gRPC method and status, Kafka topic/partition/offset, and domain IDs such as `order_id`.

HyperDX **Logs** search examples:

```text
ServiceName:orders TraceId:<hex-trace-id>
```

```text
LogAttributes.order_id:"<order-id>"
```

```text
ResourceAttributes.k8s.namespace.name:ecommerce SeverityText:ERROR
```

For a failed checkout, open the trace first; the trace side panel lists log records that share its `TraceId`. From a log record, use its trace link to open the waterfall, then narrow by `order_id` if needed. Align the time range with the trace before concluding that logs are missing.
