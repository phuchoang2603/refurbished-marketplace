## Why

The marketplace is moving toward a production-style Kubernetes platform, but internal service traffic is still opaque once requests leave the web edge. An observe-only Istio baseline gives staging and production service-to-service visibility before introducing stricter mesh security, traffic policy, or ingress changes.

## What Changes

- Add Istio as a GitOps-managed platform component for at least staging.
- Enroll the refurbished marketplace workloads in the mesh without changing application business logic.
- Correct marketplace Service port naming so Istio can classify HTTP and gRPC traffic accurately.
- Document and verify the telemetry path for request metrics, service graph, and protocol visibility against the deployed VictoriaMetrics / VictoriaTraces / Grafana stack.
- Keep mesh policy non-disruptive for this change: no strict mTLS requirement, no AuthorizationPolicy rollout, no retries/circuit breakers, and no ingress migration.

## Capabilities

### New Capabilities

- `istio-observability`: Covers observe-only Istio mesh enrollment, protocol-aware service telemetry, and rollback expectations for marketplace workloads.

### Modified Capabilities

- `argocd-gitops`: Adds GitOps-managed Istio platform installation and environment enrollment as part of the ArgoCD app-of-apps model.

## Impact

- Affects Kubernetes platform manifests under `infra/argocd/` and Helm values/templates under `infra/charts/refurbished-marketplace/` and `infra/charts/operators/istio/`.
- Adds Istio as four GitOps-managed wrapper charts (`base`, `istiod`, `cni`, `ztunnel`) for staging first, with production following the same pattern when ready.
- Changes Kubernetes Service port metadata to reflect actual protocols used by the services.
- Does not change Go service business logic, protobuf contracts, database schemas, Kafka topics, or browser-facing ingress behavior.
- Depends on the already-deployed `platform-observability` stack for Grafana / VictoriaTraces verification.
