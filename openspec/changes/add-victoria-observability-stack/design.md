## Context

Issue #1 asks for a VictoriaMetrics Kubernetes stack, Tilt wiring, service monitors, service `/metrics` endpoints, dashboards, and alerting. The Istio observe baseline now depends on an observability foundation, but does not require every Go service to expose its own Prometheus endpoint first.

The repository currently has separate Helm charts for the marketplace app, marketplace infrastructure, and Kafka. Tilt installs operators and renders those charts directly. Staging uses ArgoCD app-of-apps under `infra/argocd/staging/apps`.

The selected upstream chart is `victoriametrics/victoria-metrics-k8s-stack` from `https://victoriametrics.github.io/helm-charts/`, pinned to version `0.86.0`.

## Goals / Non-Goals

**Goals:**

- Add a repository-owned observability Helm wrapper chart around `victoria-metrics-k8s-stack`.
- Deploy the observability stack in local Tilt development and staging GitOps.
- Provide Grafana, VMSingle metrics storage, VMAgent scraping, Alertmanager, VLSingle logs storage, VLAgent log collection, VTSingle traces storage, default dashboards/rules, and a place for marketplace-owned dashboards.
- Establish the metrics, logs, and traces backend prerequisite for Istio observe mode and later application observability work.
- Document how to access Grafana and verify scrape health.

**Non-Goals:**

- No Go service `/metrics` endpoint instrumentation in this change.
- No application log shipping changes beyond deploying the cluster log backend and collector.
- No application OTLP trace emission or distributed tracing instrumentation.
- No service SLO dashboard that depends on app-specific metrics not yet emitted.
- No production rollout until staging values and storage choices are proven.

## Decisions

### Use a local wrapper chart

Create `infra/charts/observability` as a local Helm wrapper with `victoria-metrics-k8s-stack` as a dependency:

```bash
helm repo add victoriametrics https://victoriametrics.github.io/helm-charts/
helm install my-victoria-metrics-k8s-stack victoriametrics/victoria-metrics-k8s-stack --version 0.86.0
```

**Rationale:** the repo will likely own Grafana dashboards, alert rules, retention settings, and environment-specific values. A local wrapper keeps those changes reviewable with the rest of the platform.

**Alternatives considered:** ArgoCD directly referencing the upstream chart. That is simpler initially, but awkward once dashboards and custom values become repository-owned artifacts.

### Deploy in a dedicated monitoring namespace

The observability stack should deploy to a dedicated `monitoring` namespace in Tilt and staging.

**Rationale:** observability is platform infrastructure, not ecommerce application infrastructure. Separating namespaces also makes RBAC, cleanup, and port-forwarding clearer.

### Use single-node Victoria components first

Enable the single variants for every backend in this slice:

- `vmsingle` for metrics storage
- `vlsingle` for logs storage
- `vtsingle` for traces storage

Do not enable `vmcluster`, `vlcluster`, or `vtcluster` yet.

**Rationale:** this is a personal marketplace platform and staging-first rollout. Single-node components reduce operational complexity while preserving a direct path to clustered backends later if production scale requires it.

### Use default storage class with conservative retention

Do not set a chart-level storage class override. Let PersistentVolumeClaims use the cluster default storage class in local and staging environments.

Use conservative initial retention values:

- Metrics: `7d`
- Logs: `3d`
- Traces: `3d`

Use conservative initial PVC sizes:

| Backend          | Local/Tilt | Staging |
| ---------------- | ---------- | ------- |
| VMSingle metrics | `5Gi`      | `20Gi`  |
| VLSingle logs    | `5Gi`      | `20Gi`  |
| VTSingle traces  | `2Gi`      | `10Gi`  |

**Rationale:** this keeps the first rollout portable across local and staging clusters without hard-coding a storage class. Short retention limits resource use while still giving enough history to validate dashboards, alerts, and early Istio telemetry.

### Wire Tilt for local development

Tilt should create the `monitoring` namespace, render the local observability chart, and expose Grafana through `localhost:3000`.

**Rationale:** issue #1 explicitly asks for local Kubernetes observability. Local wiring helps validate chart values and dashboard packaging before staging sync.

### Add a staging ArgoCD child Application

Staging should get a dedicated observability child Application in the app-of-apps tree, ordered before Istio and before any workloads or dashboards that depend on the metrics backend.

**Rationale:** Istio observability will need metrics storage and Grafana. A separate child Application keeps this dependency visible.

### Include backends now, defer application wiring

The chart should deploy metrics, logs, and traces backends now, but this change should not wire application code to emit new metrics, logs, or traces.

**Rationale:** the chart already supports VLSingle, VLAgent, and VTSingle. Deploying the backends together gives Grafana a coherent observability surface, while keeping service changes out of this infrastructure slice.

### Prefer Istio L7 metrics for request telemetry

Once Istio observe mode is available, service request rate, latency, and error ratio should come primarily from Istio L7 telemetry rather than custom `/metrics` endpoints in every service.

**Rationale:** this avoids duplicating request instrumentation across the Go services and gives consistent cross-service telemetry. Application instrumentation should be added later only for metrics Istio cannot provide, such as business counters, queue/outbox depth, or service-local runtime details.

### Handle ArgoCD chart issues explicitly

The staging ArgoCD Application needs chart-specific sync handling:

- Add `RespectIgnoreDifferences=true` so ignored generated fields are respected during apply.
- Ignore generated VictoriaMetrics operator webhook certificate data and webhook `caBundle` drift when cert-manager is not managing those certs.
- Ignore Grafana generated admin password and the deployment checksum annotation that changes with it.
- Add `argocd.argoproj.io/sync-options: ServerSideApply=true` to default dashboard ConfigMaps to avoid large annotation failures.
- Account for ArgoCD not running Helm pre-delete hooks; avoid depending on hook-based cleanup for operator-managed resources.

**Rationale:** the upstream chart documents these ArgoCD edge cases. Capturing them in the design prevents perpetual sync drift and dashboard apply failures.

Cert-manager remains out of scope for the first rollout; self-signed VictoriaMetrics operator webhook certificates should remain ignored by ArgoCD.

## Risks / Trade-offs

- **[Chart dependency values drift]** -> Pin the VictoriaMetrics chart dependency and keep values in the local wrapper.
- **[Local stack is resource-heavy]** -> Tune local values for small retention and modest resource requests.
- **[No service `/metrics` yet]** -> Rely on Kubernetes/component metrics first, then Istio L7 metrics after mesh adoption; add app metrics only where needed.
- **[Backends exist before app signal wiring]** -> Document that logs/traces storage is deployed before app log shipping and OTLP trace emission are wired.
- **[ArgoCD sync drift from generated secrets/certs]** -> Add required ignore differences and sync options in the observability Application.
- **[Dashboards may be sparse before Istio]** -> Start with Kubernetes/platform dashboards and avoid over-promising service SLO panels until mesh metrics exist.
- **[Production storage choices are unknown]** -> Keep production rollout out of scope until staging retention and storage are validated.

## Migration Plan

1. Add the local observability wrapper chart with `victoria-metrics-k8s-stack` dependency pinned to `0.86.0`.
2. Add local values tuned for Tilt.
3. Wire Tilt to deploy the chart in `monitoring` and port-forward Grafana.
4. Add a staging ArgoCD Application for the observability chart.
5. Add staging values for single-node metrics, logs, traces, retention, storage, Grafana, Alertmanager, and default dashboards.
6. Add ArgoCD ignore differences and sync options required by the chart.
7. Add initial documentation for Grafana access and scrape target verification.
8. Validate local Tilt readiness, then staging ArgoCD sync.

## Open Questions

- None.
