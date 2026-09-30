# Marketplace observability

The shared telemetry pipeline is installed and administered by `talos-proxmox`: an OpenTelemetry agent ([`apps/components/otel-agent`](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps/components/otel-agent)) in each cluster, and one ClickHouse store with the HyperDX UI on prod ([`apps/components/observability`](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps/components/observability)). This repository owns only the marketplace telemetry producers.

## Application contract

| Signal  | Marketplace output                                                        | Platform consumer                              |
| ------- | ------------------------------------------------------------------------- | ---------------------------------------------- |
| Traces  | OTLP/gRPC to `http://otel-agent.observability.svc.cluster.local:4317`     | `otel-agent` → prod gateway → `otel_traces`    |
| Metrics | OTLP/gRPC to the same endpoint every 30 s (`OTEL_METRIC_EXPORT_INTERVAL`) | `otel-agent` → prod gateway → `otel_metrics_*` |
| Logs    | single-line slog JSON on stdout                                           | `otel-agent` filelog → `otel_logs`             |

The `otel-agent` Service uses `internalTrafficPolicy: Local`, so each pod exports to its own node's agent without credentials. The agent adds Kubernetes metadata and sets `k8s.cluster.name` and `deployment.environment` (`dev`/`prod`); the chart does not set them.

Every marketplace pod template carries `resource.opentelemetry.io/service.name: <service>`. The agent's `k8sattributes` reads it, so stdout log records get the same `ServiceName` as the service's spans and metrics. KafkaConnect pods carry `connect-debezium`.

The agent parses log bodies that start with `{` and promotes these slog keys:

| slog key   | Stored as      |
| ---------- | -------------- |
| `level`    | `SeverityText` |
| `msg`      | `Body`         |
| `trace_id` | `TraceId`      |
| `span_id`  | `SpanId`       |

All other keys (`service`, `order_id`, …) stay in `LogAttributes`. Keep those five keys stable; renaming them breaks trace-to-log navigation.

The marketplace charts render no scrape resources, dashboards, or policies for telemetry. Services open no metrics port, and marketplace CiliumNetworkPolicies are ingress-only, so OTLP egress to `observability` is not restricted. Hubble and Gateway proxy spans are outside the application telemetry path.

## Platform prerequisites

The `otel-agent` Application must be healthy in the target cluster, and prod's `observability` Application (ClickHouse, gateway, HyperDX) must be healthy for data to be stored:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get applications -n argo-cd otel-agent observability
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get pods,svc -n observability
```

Dev agents forward to prod's gateway at `10.69.12.129:4317`. When prod is down, dev services keep running and telemetry is dropped after the agents' bounded retry. Fetch kubeconfigs as described in [gitops.md](gitops.md#fresh-installation). Retention (7 days), storage, ingest authentication, and HyperDX accounts are documented in `talos-proxmox`'s [cluster access](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/operations/cluster-access.md) guide.

## HyperDX access

HyperDX is served on prod's internal LAN address `http://10.69.12.128`; it has no public hostname. It shows both environments. Filter by `ResourceAttributes['deployment.environment']` (`dev` or `prod`) when a service runs in both.

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

Expect `web`, `users`, `products`, `inventory`, `orders`, `payment`, `cart`, and `search` for all three signals. A service exports RED metrics only after it handles its first request, so exercise the shop before concluding metrics are missing. If a service is missing, check its startup logs for exporter errors and the `otel-agent` pod on the same node.

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

Save these charts to a HyperDX dashboard if you need them regularly. Dashboards live in HyperDX's own database, not in Git.

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
ServiceName:(web OR orders OR payment OR products OR inventory OR search OR cart OR users OR connect-debezium)
```

Open a `POST /checkout` span. Expect route/RPC/messaging span names, database child spans where applicable, and `connect-debezium` across asynchronous hops. Gateway proxy spans are not expected.

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
