# Observability

The observability stack is deployed from `infra/charts/observability`, a local wrapper around `victoria-metrics-k8s-stack` chart version `0.86.0`.

It deploys the first platform baseline for:

- Metrics: VMSingle and VMAgent
- Logs: VLSingle and VLAgent
- Traces: VTSingle
- Dashboards: Grafana
- Alerts: Alertmanager and stack-managed rules

Application-specific metrics endpoints, log shipping changes, and OTLP trace emission are intentionally deferred. Once Istio observe mode is available, service request rate, latency, and error ratio should prefer Istio L7 metrics where they are sufficient.

## Grafana Access

Tilt is intentionally scoped to local microservice application development and does not deploy the observability stack.

After the observability stack is deployed, open Grafana with a Kubernetes port-forward:

```bash
kubectl port-forward -n monitoring svc/vmks-grafana 3000:80
```

Then open `http://localhost:3000`.

Useful checks:

```bash
kubectl get pods -n monitoring
kubectl get svc -n monitoring
kubectl get pvc -n monitoring
```

Grafana should include datasources for VictoriaMetrics, VictoriaLogs, and VictoriaTraces. VictoriaLogs requires the `victoriametrics-logs-datasource` Grafana plugin. VictoriaTraces uses Grafana's built-in Jaeger datasource support.

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
kubectl port-forward -n monitoring svc/vmks-grafana 3000:80
```

Then open `http://localhost:3000` and confirm:

- The VictoriaMetrics datasource is present.
- The VictoriaLogs datasource is present.
- The VictoriaTraces datasource is present.
- Default Kubernetes dashboards load.
- Alertmanager is reachable from Grafana or through its service.

## ArgoCD Notes

The upstream chart has a few ArgoCD-specific behaviors that are handled in `infra/argocd/staging/apps/observability.yaml`:

- `RespectIgnoreDifferences=true` is enabled so ignored generated fields are also respected during apply.
- VictoriaMetrics Operator self-signed webhook certificate drift is ignored.
- Grafana generated admin password and related deployment checksum drift are ignored.
- Default dashboards use server-side apply to avoid large annotation failures.

ArgoCD does not run Helm pre-delete hooks, so removal should not rely on the chart's hook-based cleanup. If removing the stack, inspect operator-managed VictoriaMetrics resources in `monitoring` before deleting the namespace or Application.
