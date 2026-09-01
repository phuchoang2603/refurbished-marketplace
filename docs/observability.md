# Observability

The observability stack is deployed from `infra/charts/observability`, a local wrapper around `victoria-metrics-k8s-stack` chart version `0.86.0`.

It deploys the first platform baseline for:

- Metrics: VMSingle and VMAgent
- Logs: VLSingle and VLAgent
- Traces: VTSingle
- Dashboards: Grafana
- Alerts: Alertmanager and stack-managed rules

Application metrics endpoints and log shipping changes remain out of scope for the platform slice. Marketplace services export OTLP traces **directly** to VictoriaTraces. Cilium Hubble (L4) is the network observe path; Gateway proxies do not export traces.

### Hubble (L4)

Installed by talos-proxmox, not this chart. Dev LAN: `http://10.69.100.1`. Optional port-forward:

```bash
kubectl -n kube-system port-forward svc/hubble-ui 12000:80
```

### OTLP endpoints (direct to VTSingle)

| Protocol         | Endpoint                                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------- |
| gRPC (preferred) | `vtsingle-vmks.monitoring.svc.cluster.local:4317` (insecure TLS in staging)              |
| HTTP fallback    | `http://vtsingle-vmks.monitoring.svc.cluster.local:10428/insert/opentelemetry/v1/traces` |

Set service env `OTEL_EXPORTER_OTLP_ENDPOINT=vtsingle-vmks.monitoring.svc.cluster.local:4317` (gRPC) or use the HTTP URL with the `shared/observe/trace` bootstrap’s HTTP mode.

## Grafana Access

Argo CD deploys the observability chart into `monitoring`. Cilium Gateway + HTTPRoute (Cloudflare origin):

- Dev: `https://grafana-dev.phuchoang.sbs`
- Prod: `https://grafana.phuchoang.sbs`

Point the tunnel Public Hostname at `http://cilium-gateway-grafana.monitoring.svc.cluster.local:80`.

Port-forward Grafana:

```bash
kubectl port-forward -n monitoring svc/observability-grafana 3000:80
```

Open the public hostname (or http://localhost:3000) and sign in:

- **Username:** `admin`
- **Password:** generated into Secret `observability-grafana` (key `admin-password`)

```bash
kubectl get secret observability-grafana -n monitoring \
  -o jsonpath='{.data.admin-password}' | base64 -d && echo
```

Useful checks:

```bash
kubectl get pods -n monitoring
kubectl get svc -n monitoring
kubectl get pvc -n monitoring
```

Grafana should include datasources for VictoriaMetrics, VictoriaLogs, and VictoriaTraces. VictoriaLogs requires the `victoriametrics-logs-datasource` Grafana plugin. VictoriaTraces is provisioned as a Grafana **Tempo** datasource (`http://vtsingle-vmks.monitoring.svc.cluster.local:10428/select/tempo`) so Explore can use TraceQL and (optionally) Grafana Traces Drilldown.

Default dashboards are fetched by `vmks-sync-job` (an Argo CD `PostSync` hook in the wrapper chart) and loaded into Grafana via the dashboard sidecar. After sync, verify:

```bash
kubectl get configmaps -n monitoring -l grafana_dashboard=1
```

## Staging Health Checks

After ArgoCD syncs `staging-observability`, check the Application and namespace:

```bash
kubectl get applications.argoproj.io -n argo-cd staging-observability
kubectl get pods -n monitoring
kubectl get pvc -n monitoring
```

Check VictoriaMetrics Operator custom resources:

```bash
kubectl get vmsingle,vlagent,vlsingle,vtsingle,vmagent,vmalert -n monitoring
```

Check Grafana, Alertmanager, and service endpoints:

```bash
kubectl get svc -n monitoring
```

When Grafana access is available, confirm:

- The VictoriaMetrics datasource is present.
- The VictoriaLogs datasource is present.
- The VictoriaTraces datasource is present.
- Default Kubernetes dashboards load.
- Alertmanager is reachable from Grafana or through its service.

## ArgoCD Notes

The upstream chart has a few ArgoCD-specific behaviors that are handled on the observability Application in `infra/argocd/app-of-apps/templates/applications.tpl`:

- `managedNamespaceMetadata` labels `monitoring` `pod-security.kubernetes.io/enforce=privileged` so Talos default PSS baseline does not block node-exporter.
- `RespectIgnoreDifferences=true` is enabled so ignored generated fields are also respected during apply.
- VictoriaMetrics Operator self-signed webhook certificate drift is ignored.
- Grafana generated admin password and related deployment checksum drift are ignored.
- Default dashboards use server-side apply to avoid large annotation failures.
- The upstream `vmks-sync-job` Helm hook is disabled; the wrapper chart runs an equivalent Argo CD `PostSync` job so dashboards are provisioned on sync.

ArgoCD does not run Helm pre-delete hooks, so removal should not rely on the chart's hook-based cleanup. If removing the stack, inspect operator-managed VictoriaMetrics resources in `monitoring` before deleting the namespace or Application.

## Distributed tracing (e2e)

App tracing bootstrap lives in `shared/observe/trace` (wired through `shared/runtime`). Marketplace services export OTLP **directly** to VictoriaTraces (no collector). Istio ambient stays enrolled for L7 **metrics** only — mesh proxy tracing (Gateway Telemetry / Envoy OTEL) is not used.

```
Browser → ingress → web ──gRPC──▶ domain services (+ DB / Redis child spans)
                         │
                    outbox.tracingspancontext
                         │
              Debezium EventRouter (+ Strimzi OTEL agent on Connect)
                         │  Kafka header traceparent
                         ▼
              consumers (child-of spans) → VictoriaTraces → Grafana Explore
```

**Joining rule:** one W3C `TraceId` across app spans when services propagate `traceparent`. Async hops continue via the outbox column → Kafka headers. Consumer spans use parent–child (not links) for Grafana waterfall UX.

**Span naming:** Prefer operation names over process names. Web HTTP server spans use `METHOD` + chi route pattern (e.g. `POST /orders/{id}`, `http.route` set). gRPC uses the full method; Kafka consumers use `messaging process <topic>`. Postgres (otelsql via `OpenPostgres`) and Redis (redisotel via `OpenRedis`) appear as child spans under those parents. Bound SQL parameter values are not recorded; statement text is truncated.

**Connect tracing:** KafkaConnect sets `spec.tracing.type: opentelemetry` (loads Strimzi `tracing-agent`) plus `OTEL_PROPAGATORS=tracecontext` and OTLP export to VictoriaTraces. EventRouter maps `tracingspancontext` → Kafka `traceparent`. Rebuild `connect-debezium` only when the Debezium plugin changes; enabling the agent is a chart/CR change.

**Mesh:** Keep ambient + waypoint/ingress **metrics** (Marketplace Istio RED). Trace waterfalls are app + Connect only — no `ecommerce-ingress` / `ecommerce-waypoint` spans.

**Verify after deploy:**

1. Confirm VT Service has port `4317` and apps have `OTEL_EXPORTER_OTLP_ENDPOINT`.
2. Place a checkout order; in Grafana Explore select the **VictoriaTraces** Tempo datasource. Prefer TraceQL scoped to app services:

```
{ resource.service.name =~ "web|orders|payment|products|cart|users|connect-debezium" }
```

3. Open a TraceId for service `web`: root should look like `POST /cart/checkout` (or similar route pattern), not the bare string `web`. Expect web → orders (gRPC + DB children) → Debezium/connect → products (messaging + DB). Kafka `orders.created` records should carry a `traceparent` header.
4. Complete hosted-payment success/fail; confirm callback → payment gRPC → payment outbox path.
5. Confirm mesh proxy spans are absent. Edge SLIs stay on **Marketplace Istio RED** (Istio L7 metrics), not VictoriaTraces.

## Structured logging

Marketplace services emit **JSON slog** lines to stdout via `shared/observe/log` (wired by `shared/runtime.InitLogging`). Call sites use that package’s helpers — prefer `InfoContext` / `WarnContext` / `ErrorContext` on request paths so `trace_id` / `span_id` are injected; use `Key*` constants with key/value pairs (or `Attr*` with `LogAttrs`). Do not use raw `log/slog`. VLAgent scrapes those lines into VictoriaLogs.

VLAgent scrapes the `ecommerce` namespace (apps + CNPG DB pods) and skips `wait-for-db` init containers. Use Hubble for L4 flows.

HTTP/gRPC access logs put the useful bits in `msg` (e.g. `GET /orders/... 200`, `ListOrdersByBuyer OK`) while keeping structured attrs for filters. Log `level` is emitted lowercase (`info`, `error`, …) so Grafana Explore does not mark marketplace JSON as `unknown`.

### Field conventions

| Field                                        | When present                        |
| -------------------------------------------- | ----------------------------------- |
| `service`                                    | Always (bootstrap default)          |
| `trace_id`                                   | When logging with a valid OTEL span |
| `span_id`                                    | When logging with a valid OTEL span |
| `method`/`path`/`status`/`duration_ms`       | HTTP access logs (web)              |
| `grpc_method`/`grpc_code`/`duration_ms`      | gRPC unary access logs              |
| `topic`/`partition`/`offset`                 | Kafka handler errors                |
| `order_id` / `merchant_id` / `buyer_user_id` | Checkout hot-path domain logs       |
| `status` / `outcome` / `event_type`          | Order/payment/reservation outcomes  |

Sensitive attribute keys (`password`, `token`, `api_key`, `bearer`, `access_token`, `card`, `cvv`, …) are redacted to `[redacted]` via `ReplaceAttr`. Redaction does not rewrite free-text `msg` — never put secrets in the message string. Do not log full payment gateway payloads.

### LogSQL examples (VictoriaLogs)

```logsql
service:="web"
```

```logsql
service:="orders" AND trace_id:="<hex-trace-id>"
```

```logsql
order_id:="<uuid>"
```

Exact filter syntax can vary slightly with the Grafana VL plugin UI — prefer Explore’s builder, then copy the query.

### Debug a checkout

Logs Drilldown does not work with VictoriaLogs — use **Explore**.

1. **Traces:** Explore → **VictoriaTraces** → TraceQL `{ resource.service.name =~ "web|orders|payment|products|cart|users|connect-debezium" }` (or search service `web`) → open a TraceId. Expect app spans with route/RPC/messaging names, DB/Redis children where applicable, and `connect-debezium` on the async hop — not mesh proxy services.
2. **App logs for that TraceId:** Trace → logs (Tempo `tracesToLogsV2` filters by `trace_id` only). You should see marketplace JSON lines across services for that TraceId.
3. **App logs in Explore** (same time range as the trace):

```logsql
kubernetes.pod_namespace:="ecommerce"
  AND service:in(web,orders,payment,products,cart,users)
```

4. **Optional drills:** `order_id:="<uuid>"` on app JSON; for the Kafka hop use TraceId spans (`connect-debezium`) rather than Connect pod logs (not scraped into VL). Edge latency/errors: **Dashboards → Marketplace → Marketplace Istio RED** (Istio L7 metrics).

### Trace → logs

The VictoriaTraces Tempo datasource (`uid: VictoriaTraces`) is provisioned with `tracesToLogsV2` pointing at VictoriaLogs (`uid: VictoriaLogs`, `filterByTraceID: true`, no `service.name` tag filter).

1. Open a span in Explore / Traces Drilldown.
2. Use **Logs for this span** / Trace → logs.
3. Confirm matching JSON lines include the same `trace_id` (all marketplace services that logged for that TraceId).
