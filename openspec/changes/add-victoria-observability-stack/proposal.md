## Why

The marketplace needs a centralized observability foundation before Istio observability and service SLO work can be trusted. A VictoriaMetrics Kubernetes stack provides metrics, logs, traces, dashboards, and alerting backends while application-level instrumentation can still be added later.

## What Changes

- Add a local Helm wrapper chart for `victoria-metrics-k8s-stack` from `https://victoriametrics.github.io/helm-charts/`, pinned to version `0.86.0`, so repository-owned values, dashboards, and alert rules can grow over time.
- Deploy the observability stack through the existing ArgoCD app-of-apps model for staging.
- Keep Tilt focused on microservice application development; do not deploy the observability stack from `Tiltfile`.
- Include Grafana, VMSingle metrics storage, VMAgent scraping, Alertmanager, VLSingle logs storage, VLAgent log collection, VTSingle traces storage, default Kubernetes dashboards, and room for repository-owned dashboards.
- Treat Istio L7 request metrics as the preferred source for service request rate, latency, and error telemetry once Istio is available.
- Defer application `/metrics` instrumentation unless a metric cannot be obtained from Kubernetes, VictoriaMetrics stack components, Kafka/Connect exporters, or Istio.
- Deploy logs and traces backends in this slice, but defer wiring application log shipping, OTLP trace emission, and service instrumentation.
- Include ArgoCD-specific sync and ignore-difference handling for VictoriaMetrics operator webhooks, Grafana generated secrets, and large dashboard ConfigMaps.

## Capabilities

### New Capabilities

- `platform-observability`: Covers the VictoriaMetrics Kubernetes metrics, logs, traces, dashboards, alerting stack, Grafana access, local Helm validation, and staging GitOps deployment.

### Modified Capabilities

- `argocd-gitops`: Adds a staging observability Application to the app-of-apps model.

## Impact

- Affects `infra/charts/` by adding an observability wrapper chart.
- Affects `infra/argocd/staging/apps/` by adding a staging observability application.
- Affects docs by adding operator/developer guidance for opening Grafana and validating scrape health.
- Does not affect `Tiltfile`; Tilt remains scoped to developing the marketplace microservice applications.
- Does not add Go service instrumentation, service `/metrics` endpoints, application log shipping, or application trace emission in this change.
