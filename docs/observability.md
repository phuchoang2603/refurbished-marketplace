# Observability

The observability stack is deployed from `infra/charts/observability`, a local wrapper around `victoria-metrics-k8s-stack` chart version `0.86.0`.

It deploys the first platform baseline for:

- Metrics: VMSingle and VMAgent
- Logs: VLSingle and VLAgent
- Traces: VTSingle
- Dashboards: Grafana
- Alerts: Alertmanager and stack-managed rules

Application-specific metrics endpoints, log shipping changes, and OTLP trace emission are intentionally deferred. Once Istio observe mode is available, service request rate, latency, and error ratio should prefer Istio L7 metrics where they are sufficient.

The observability wrapper scrapes Istio ambient components into VMSingle via `VMPodScrape` resources (`istioScrapes.enabled`):

| Target    | Namespace                  | Port                          |
| --------- | -------------------------- | ----------------------------- |
| istiod    | `istio-system`             | `http-monitoring` (`15014`)   |
| ztunnel   | `istio-system`             | `ztunnel-stats` (`15020`)     |
| istio-cni | `istio-system`             | `metrics` (`15014`)           |
| waypoint  | `ecommerce` (configurable) | `http-envoy-prom` / `metrics` |

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
