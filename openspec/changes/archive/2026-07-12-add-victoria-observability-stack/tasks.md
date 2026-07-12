## 1. Chart Wrapper

- [x] 1.1 Add the `victoriametrics` Helm repository source `https://victoriametrics.github.io/helm-charts/`.
- [x] 1.2 Add `infra/charts/observability/Chart.yaml` with `victoria-metrics-k8s-stack` pinned to version `0.86.0`.
- [x] 1.3 Add default chart values for local/dev observability.
- [x] 1.4 Add staging values for VMSingle metrics storage, VLSingle logs storage, VTSingle traces storage, retention, storage, Grafana, Alertmanager, and default dashboards.
- [x] 1.5 Add a repository-owned dashboard/rules location under the observability chart for future Grafana dashboards and alert rules.
- [x] 1.6 Ensure VMCluster, VLCluster, and VTCluster are not required for this slice.
- [x] 1.7 Leave storage class unset so PVCs use the cluster default storage class.
- [x] 1.8 Configure initial retention as metrics `7d`, logs `3d`, and traces `3d`.
- [x] 1.9 Configure local PVC sizes as VMSingle `5Gi`, VLSingle `5Gi`, and VTSingle `2Gi`.
- [x] 1.10 Configure staging PVC sizes as VMSingle `20Gi`, VLSingle `20Gi`, and VTSingle `10Gi`.

## 2. Tilt Scope

- [x] 2.1 Keep `Tiltfile` scoped to marketplace microservice application development.
- [x] 2.2 Do not create or use the `monitoring` namespace from `Tiltfile`.
- [x] 2.3 Do not render the local observability chart from `Tiltfile`.
- [x] 2.4 Do not add a Tilt-managed Grafana port-forward.

## 3. Staging GitOps

- [x] 3.1 Add a staging ArgoCD Application for the observability chart.
- [x] 3.2 Configure the staging Application to deploy into the `monitoring` namespace.
- [x] 3.3 Set observability sync ordering before Istio dashboard or telemetry verification depends on it.
- [x] 3.4 Add `RespectIgnoreDifferences=true` to the observability Application sync options.
- [x] 3.5 Add ignore differences for VictoriaMetrics operator validation Secret data and webhook `caBundle` drift.
- [x] 3.6 Add ignore differences for Grafana generated admin password Secret data and related deployment checksum annotation drift.
- [x] 3.7 Add server-side apply handling for default dashboard ConfigMaps.
- [x] 3.8 Avoid relying on Helm pre-delete hooks for ArgoCD cleanup behavior.
- [x] 3.9 Keep production observability deployment out of scope for this change.
- [x] 3.10 Keep self-signed VictoriaMetrics operator webhook certificates ignored by ArgoCD; do not introduce cert-manager in this change.

## 4. Observability Backend Baseline

- [x] 4.1 Confirm Grafana has a VictoriaMetrics-compatible datasource.
- [x] 4.2 Confirm Grafana has a VictoriaLogs datasource and required logs plugin.
- [x] 4.3 Confirm Grafana has a VictoriaTraces datasource via the Jaeger-compatible API.
- [x] 4.4 Confirm VMAgent or stack-managed scraping is active for Kubernetes/platform targets.
- [x] 4.5 Confirm VLAgent or stack-managed log collection resources are deployed.
- [x] 4.6 Confirm VTSingle trace storage resources are deployed.
- [x] 4.7 Confirm Alertmanager deploys and can receive stack alert rules.
- [x] 4.8 Avoid adding Go service `/metrics` endpoints in this change.
- [x] 4.9 Avoid wiring application log shipping or OTLP trace emission in this change.
- [x] 4.10 Document that Istio L7 metrics are the preferred later source for request rate, latency, and error ratio where sufficient.

## 5. Documentation And Validation

- [x] 5.1 Add observability documentation for Grafana access.
- [x] 5.2 Add observability documentation for staging scrape health, log backend, and trace backend checks.
- [x] 5.3 Validate Helm rendering for the observability chart.
- [x] 5.4 Run OpenSpec validation for `add-victoria-observability-stack`.
- [x] 5.5 Update GitHub issue #1 with the revised backend-first scope, chart version, ArgoCD caveats, and follow-ups.
