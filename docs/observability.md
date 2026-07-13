# Observability

The observability stack is deployed from `infra/charts/observability`, a local wrapper around `victoria-metrics-k8s-stack` chart version `0.86.0`.

It deploys the first platform baseline for:

- Metrics: VMSingle and VMAgent
- Logs: VLSingle and VLAgent
- Traces: VTSingle
- Dashboards: Grafana
- Alerts: Alertmanager and stack-managed rules

Application metrics endpoints and log shipping changes remain out of scope for the platform slice. Marketplace services and Istio export OTLP traces **directly** to VictoriaTraces (no OpenTelemetry Collector required).

The observability wrapper scrapes Istio L7 proxies into VMSingle via `VMPodScrape` resources (`istioScrapes.enabled`):

| Target   | Namespace                  | Port                          |
| -------- | -------------------------- | ----------------------------- |
| waypoint | `ecommerce` (configurable) | `http-envoy-prom` / `metrics` |
| ingress  | `ecommerce` (configurable) | `http-envoy-prom` / `metrics` |

Control-plane targets (istiod, ztunnel, istio-cni) are intentionally not scraped.

### OTLP endpoints (direct to VTSingle)

| Protocol         | Endpoint                                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------- |
| gRPC (preferred) | `vtsingle-vmks.monitoring.svc.cluster.local:4317` (insecure TLS in staging)              |
| HTTP fallback    | `http://vtsingle-vmks.monitoring.svc.cluster.local:10428/insert/opentelemetry/v1/traces` |

Set service env `OTEL_EXPORTER_OTLP_ENDPOINT=vtsingle-vmks.monitoring.svc.cluster.local:4317` (gRPC) or use the HTTP URL with the shared bootstrap’s HTTP mode.

Useful PromQL after marketplace traffic:

```promql
sum by (destination_app, request_protocol) (
  rate(istio_requests_total{destination_service_namespace="ecommerce"}[5m])
)
```

## Grafana Access

Local and staging Argo CD both deploy the observability chart into `monitoring` (local uses chart `values.yaml`; staging uses `values-staging.yaml`).

Port-forward Grafana:

```bash
kubectl port-forward -n monitoring svc/observability-grafana 3000:80
```

Open http://localhost:3000 and sign in:

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

Grafana should include datasources for VictoriaMetrics, VictoriaLogs, and VictoriaTraces. VictoriaLogs requires the `victoriametrics-logs-datasource` Grafana plugin. VictoriaTraces uses Grafana's built-in Jaeger datasource support.

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

- `RespectIgnoreDifferences=true` is enabled so ignored generated fields are also respected during apply.
- VictoriaMetrics Operator self-signed webhook certificate drift is ignored.
- Grafana generated admin password and related deployment checksum drift are ignored.
- Default dashboards use server-side apply to avoid large annotation failures.
- The upstream `vmks-sync-job` Helm hook is disabled; the wrapper chart runs an equivalent Argo CD `PostSync` job so dashboards are provisioned on sync.

ArgoCD does not run Helm pre-delete hooks, so removal should not rely on the chart's hook-based cleanup. If removing the stack, inspect operator-managed VictoriaMetrics resources in `monitoring` before deleting the namespace or Application.

## Distributed tracing (e2e)

Marketplace services and Istio export OTLP **directly** to VictoriaTraces (no collector).

```
Browser → ingress → web ──gRPC──▶ domain services
                         │
                    outbox.tracingspancontext
                         │
              Debezium EventRouter (+ OTEL on Connect)
                         │  Kafka header traceparent
                         ▼
              consumers (child-of spans) → VictoriaTraces → Grafana Explore
```

**Joining rule:** one W3C `TraceId` across app spans and Istio L7 hops when apps propagate `traceparent`. Async hops continue via the outbox column → Kafka headers. Consumer spans use parent–child (not links) for Grafana waterfall UX.

**Verify after deploy:**

1. Confirm VT Service has port `4317` and apps have `OTEL_EXPORTER_OTLP_ENDPOINT`.
2. Place a checkout order; in Grafana Explore (VictoriaTraces) search by service `web` / recent traces.
3. Confirm the TraceId includes web → orders → Debezium/connect → products (inventory) spans.
4. Complete hosted-payment success/fail; confirm callback → payment → payment outbox path.
5. Confirm mesh spans appear when Istio Telemetry `ecommerce-tracing` is enabled.
