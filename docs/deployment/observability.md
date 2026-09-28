# Marketplace observability

The shared VictoriaMetrics, VictoriaLogs, VictoriaTraces, Grafana, and Alertmanager stack is installed and administered by [`talos-proxmox/apps/components/observability`](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps/components/observability). This repository owns the marketplace telemetry producers, scrape resource, network access, and two Grafana dashboard ConfigMaps.

## Application contract

| Signal     | Marketplace output                                                                                             | Platform consumer         |
| ---------- | -------------------------------------------------------------------------------------------------------------- | ------------------------- |
| Metrics    | Prometheus `/metrics` on pod port `9100`; `VMPodScrape/marketplace-apps` selects `marketplace.metrics: "true"` | VMAgent → VictoriaMetrics |
| Logs       | structured JSON on stdout                                                                                      | VLAgent → VictoriaLogs    |
| Traces     | OTLP/gRPC to `vtsingle-vmks.monitoring.svc.cluster.local:4317`                                                 | VictoriaTraces            |
| Dashboards | ConfigMaps in `monitoring` labeled `grafana_dashboard: "1"`                                                    | Grafana dashboard sidecar |

The `refurbished-marketplace` Helm release provisions **Marketplace RED** and **Marketplace logs**. Grafana folder placement is controlled by the platform sidecar/provider configuration. The marketplace release does not create `monitoring`, Grafana, datasources, collectors, or storage.

Hubble and Gateway proxy spans are outside the application RED/tracing path. Metrics use scrape, while traces use direct OTLP export.

## Platform prerequisites

The environment's `observability` platform Application provides the telemetry services. Marketplace children retry until its CRDs exist; to check the stack directly:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get applications -n argo-cd
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get pods,svc -n monitoring
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get \
  vmsingle,vlagent,vlsingle,vtsingle,vmagent -n monitoring
```

Fetch kubeconfigs as described in [gitops.md](gitops.md#fresh-installation). Platform installation, retention, storage, datasource, ingress, and credential details live in `talos-proxmox`'s [observability component](https://github.com/phuchoang2603/talos-proxmox/tree/main/apps/components/observability) and [GitOps architecture](https://github.com/phuchoang2603/talos-proxmox/blob/main/docs/architecture/gitops.md) guide.

## Grafana access

- Dev: `https://grafana-dev.phuchoang.sbs`
- Prod: `https://grafana.phuchoang.sbs`

For local access:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" \
  port-forward -n monitoring svc/observability-grafana 3000:80
```

The platform provides datasources with UIDs `VictoriaMetrics`, `VictoriaLogs`, and `VictoriaTraces`. Confirm marketplace dashboards are discovered:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" \
  get configmaps -n monitoring -l grafana_dashboard=1
```

## Verify application telemetry

Confirm the scrape resource and exporter configuration:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get vmpodscrape marketplace-apps -n ecommerce
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get pods -n ecommerce -l marketplace.metrics=true
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get deploy -n ecommerce \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[0].env[?(@.name=="OTEL_EXPORTER_OTLP_ENDPOINT")].value}{"\n"}{end}'
```

To check successful scraping, query VictoriaMetrics in Grafana Explore:

```promql
up{namespace="ecommerce",job=~"web|users|products|inventory|orders|payment|cart|search"}
```

Expect one series per metrics-enabled pod, each with value `1`. A `0` means the target was discovered but scraping failed; absent series mean discovery or collection needs investigation. Compare results with the metrics-enabled pod list above.

For target errors, port-forward VMAgent and open `http://localhost:8429/targets`:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" \
  port-forward -n monitoring svc/vmagent-vmks 8429:8429
```

Find the `ecommerce` targets on port `9100` with path `/metrics`, confirm they are UP and have a recent successful scrape, and inspect Last Error for failed targets. Check the pod endpoint, VMPodScrape selection, and scrape network policy when targets fail.

In Grafana:

- use **Marketplace RED** for HTTP/gRPC request rate, errors, and p95 latency;
- use **Marketplace logs** or VictoriaLogs Explore for JSON logs;
- use the `VictoriaTraces` Tempo datasource for TraceQL and waterfalls.

## Distributed tracing

Application tracing is initialized through `shared/observe/trace` and `shared/runtime`. Services propagate W3C `traceparent`; outbox rows carry `tracingspancontext`, Debezium maps it to Kafka headers, and consumers continue the trace.

```text
Browser → web → domain services → database
                    │
                    └→ outbox → Debezium/Connect → Kafka → consumers
                                                              │
                                                              └→ VictoriaTraces
```

KafkaConnect enables Strimzi OpenTelemetry tracing and exports to the same VictoriaTraces endpoint. Rebuild the Connect image only when its plugin contents change.

To verify a checkout, query:

```traceql
{ resource.service.name =~ "web|orders|payment|products|inventory|search|cart|users|connect-debezium" }
```

Expect route/RPC/messaging span names, database child spans where applicable, and `connect-debezium` across asynchronous hops. Gateway proxy spans are not expected.

## Structured logging and correlation

Marketplace services emit JSON slog lines via `shared/observe/log`. Request-path logging should use context-aware helpers so `trace_id` and `span_id` are present. Sensitive keys are redacted, but free-text messages are not rewritten; never put secrets or payment payloads in log messages.

Common fields include `service`, `trace_id`, `span_id`, HTTP/gRPC method and status, Kafka topic/partition/offset, and domain IDs such as `order_id`.

VictoriaLogs examples:

```logsql
service:="orders" AND trace_id:="<hex-trace-id>"
```

```logsql
kubernetes.pod_namespace:="ecommerce"
  AND service:in(web,orders,payment,products,inventory,search,cart,users)
```

The platform `VictoriaTraces` datasource sends Trace → logs queries as LogsQL using `trace_id`; it does not use Loki stream selectors. The platform `VictoriaLogs` datasource provides derived links from `trace_id` to traces and from `service` to VictoriaMetrics.

For a failed checkout, open the trace first, follow Trace → logs, then narrow by `order_id` if available. Align the Grafana time range with the trace before concluding that logs are missing.
