## Context

The remote Kubernetes delivery model uses ArgoCD app-of-apps. Staging currently syncs operator applications, marketplace infrastructure, the marketplace Helm chart, and Kafka from `infra/argocd/staging/apps`. The marketplace services communicate internally through Kubernetes Services and gRPC, with Kafka and database dependencies managed separately.

The marketplace chart currently renders one `Deployment` and `Service` per enabled service. Every Service port is named `http`, even though the backend service ports are primarily gRPC. That naming is acceptable for basic Kubernetes service discovery, but it weakens Istio protocol detection and the quality of generated telemetry.

## Goals / Non-Goals

**Goals:**

- Install Istio through the existing ArgoCD/GitOps model for staging first.
- Enroll marketplace workloads in the mesh with an observe-only posture.
- Make service port naming protocol-aware so HTTP and gRPC traffic are classified correctly.
- Verify mesh telemetry for internal marketplace service traffic.
- Preserve existing Tilt/local development behavior.

**Non-Goals:**

- No ingress or edge routing migration.
- No strict mTLS enforcement.
- No AuthorizationPolicy rollout.
- No retries, circuit breakers, traffic splitting, mirroring, or canary policy.
- No application business logic changes.
- No Kafka, CNPG, ESO, or database platform replacement.

## Decisions

### Use staging as the first mesh environment

Staging gets the first GitOps-managed Istio installation and workload enrollment. Production follows only after staging proves that sync, enrollment, telemetry, and rollback are understood.

**Alternatives considered:** enabling staging and production together. Rejected because mesh adoption changes pod networking and observability assumptions; staging should absorb that risk first.

### Head toward Istio ambient mode with waypoint proxy

The target direction is Istio ambient mode, with waypoint proxies introduced for workloads that need L7 policy, telemetry, or routing behavior. The observe-only baseline should prefer ambient enrollment if the staging cluster prerequisites are satisfied.

**Rationale:** ambient mode reduces application pod sidecar coupling and keeps mesh infrastructure more separate from application development. Waypoints provide a path to L7 behavior without requiring every workload to carry an Envoy sidecar.

**Alternatives considered:** sidecar injection as the default. It is mature and widely documented, but it changes each pod's container shape and resource profile more directly. Keep it as a fallback if ambient prerequisites are not met.

### Pin Istio to the official 1.30.2 Helm release

Istio installation should use the official Helm repository:

```bash
helm repo add istio-official https://istio-release.storage.googleapis.com/charts
```

The first pinned control plane target is `istio-official/istiod` at version `1.30.2`. Ambient-mode installation also needs the matching Istio base, CNI, and ztunnel components required by the selected installation profile.

**Rationale:** pinning the upstream release keeps staging reproducible and avoids floating chart behavior during GitOps syncs.

### Keep the first rollout observe-only

The first Istio change will focus on installation, enrollment, protocol detection, and telemetry. It will not add strict mesh policy or traffic management rules.

**Rationale:** the current services were not written with mesh-enforced identity or retries in mind. Observability gives immediate value and produces evidence for later mTLS, authorization, and routing work.

### Prefer GitOps-managed Istio platform resources

Istio installation should be represented as ArgoCD-managed manifests or Helm Applications in the environment app-of-apps tree, matching the existing operator pattern.

**Rationale:** remote clusters already use ArgoCD as the control surface. Keeping Istio in that model makes drift, rollback, and environment promotion visible in Git.

### Make mesh enrollment explicit and reversible

Marketplace namespace or workload enrollment should be controlled through chart values or environment-specific manifests rather than implicit cluster-wide defaults.

**Rationale:** explicit enrollment makes rollback straightforward and avoids surprising local/Tilt behavior. If sidecar mode is selected, namespace injection labels can be managed by GitOps. If ambient mode is selected, enrollment labels can follow the same principle.

### Classify service ports by actual protocol

The marketplace chart should render port names that match the service protocol. The web and payment gateway simulator services are HTTP. Internal gRPC services such as users, products, orders, cart, and payment should expose `grpc` port names.

**Rationale:** Istio uses service metadata for protocol selection and telemetry. Correct names help ensure gRPC requests appear as gRPC rather than generic HTTP or opaque TCP.

### Use VictoriaTraces with Grafana as the telemetry direction

The target observability stack is VictoriaTraces with Grafana, alongside the VictoriaMetrics Kubernetes stack. This stack should be proposed as a separate prerequisite platform observability change. Istio implementation should avoid assuming a different long-term tracing UI, and should document which verification checks are blocked until the prerequisite observability stack exists.

**Rationale:** Istio can emit telemetry before the final visualization stack is complete, but request tracing and dashboards need a backend. Aligning on VictoriaTraces/Grafana avoids building temporary documentation around another tracing tool.

**Dependency note:** the VictoriaMetrics/VictoriaTraces proposal should be created before implementation planning for Istio proceeds. It does not have to be fully deployed before every Istio manifest is written, but it must exist as the prerequisite plan for trace and dashboard verification.

## Risks / Trade-offs

- **[Istio adds platform complexity]** -> Limit the first change to staging and observe-only behavior.
- **[Protocol naming mistakes reduce telemetry quality]** -> Make protocol an explicit service value and verify rendered Services before rollout.
- **[Sidecar injection can change pod startup/resource behavior]** -> Keep enrollment reversible and verify main browser flows after sync.
- **[Kafka and database traffic may not produce useful L7 telemetry]** -> Treat this change as service request observability first; document lower-layer dependencies separately.
- **[Production rollout may need environment-specific choices]** -> Do not enable production until staging results are known.
- **[Telemetry stack may lag mesh install]** -> Create a separate VictoriaMetrics/VictoriaTraces prerequisite proposal and keep trace/dashboard checks blocked on it.

## Migration Plan

1. Create the separate VictoriaMetrics/VictoriaTraces prerequisite proposal.
2. Verify staging cluster CNI, node, Kubernetes, and Gateway API readiness for Istio ambient mode.
3. Add a staging ArgoCD application for Istio installation using the official Istio Helm repository and pinned `1.30.2` release.
4. Add explicit marketplace mesh enrollment for staging.
5. Sync staging and verify Istio control plane, CNI, and ztunnel health.
6. Exercise web, product, cart, checkout, and payment flows.
7. Confirm mesh telemetry shows internal service traffic and expected protocols.
8. Document rollback: remove marketplace enrollment first, then remove or disable the Istio application if needed.

## Open Questions

- Are the staging cluster CNI and node capabilities ready for Istio ambient mode?
- Which Istio ambient components should be split into separate ArgoCD Applications versus one grouped Istio platform application?
