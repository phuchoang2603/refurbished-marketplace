## Context

The staging ArgoCD app-of-apps already deploys operators, the VictoriaMetrics observability stack (`monitoring`), the marketplace Helm chart, and Kafka. The marketplace services communicate internally through Kubernetes Services and gRPC, with Kafka and database dependencies managed separately.

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

Istio installation should use local wrappers that depend on the official Helm repository charts:

```bash
helm repo add istio https://istio-release.storage.googleapis.com/charts
```

Each wrapper under `infra/charts/operators/istio/` pins its upstream chart to `1.30.2`. Ambient mode requires `istiod` and `cni` with `profile=ambient`, plus `base` and `ztunnel`.

**Rationale:** pinning the upstream release keeps staging reproducible and avoids floating chart behavior during GitOps syncs. Wrappers keep values and pins reviewable in-repo like the other operator charts.

### Keep the first rollout observe-only

The first Istio change will focus on installation, enrollment, protocol detection, and telemetry. It will not add strict mesh policy or traffic management rules.

**Rationale:** the current services were not written with mesh-enforced identity or retries in mind. Observability gives immediate value and produces evidence for later mTLS, authorization, and routing work.

### Prefer GitOps-managed Istio platform resources

Istio installation should be represented as ArgoCD-managed Helm Applications in the environment app-of-apps tree, matching the existing operator pattern.

**Rationale:** remote clusters already use ArgoCD as the control surface. Keeping Istio in that model makes drift, rollback, and environment promotion visible in Git.

### Use four Istio wrapper charts under `infra/charts/operators/istio`

Ship four local Helm wrappers that each pin one official Istio chart to `1.30.2`:

```
infra/charts/operators/istio/
├── base/      # istio/base
├── istiod/    # istio/istiod (profile=ambient)
├── cni/       # istio/cni (profile=ambient)
└── ztunnel/   # istio/ztunnel
```

Staging gets four corresponding ArgoCD Applications (for example `staging-istio-base`, `staging-istiod`, `staging-istio-cni`, `staging-ztunnel`) with sync waves enforcing:

1. `base` (CRDs)
2. `istiod` + `cni` (ambient profile)
3. `ztunnel`
4. marketplace enrollment (after Istio is healthy)

**Rationale:** the official ambient Helm install is already four charts with a required order. Separate Applications match the ESO/CNPG/Strimzi pattern and allow independent upgrade/rollback. Grouping the wrappers under one `istio/` folder keeps them discoverable without collapsing them into a single Helm release.

**Alternatives considered:** one umbrella Istio chart like observability. Rejected because Istio’s upstream packaging is modular and CRD upgrades should not be forced with every ztunnel bump.

### Confirm staging is ambient-capable (with Canal caveats)

Staging (`dev-rke2`) was checked against ambient prerequisites:

| Check            | Result                                                                          |
| ---------------- | ------------------------------------------------------------------------------- |
| Kubernetes       | `v1.32.3+rke2r1` — supported for Istio 1.30                                     |
| Nodes            | Ubuntu 24.04, kernel 6.8, amd64, containerd — fine                              |
| Cluster CNI      | RKE2 Canal (Calico + Flannel) — supported ambient path since in-pod redirection |
| Gateway API CRDs | Already installed — good for later waypoints                                    |

Proceed with ambient as the target. During install, verify `istio-cni` and `ztunnel` health. Allow TCP **15008** (HBONE) in any NetworkPolicy that would otherwise block mesh traffic. Fall back to sidecar only if ambient dataplane redirect fails on Canal.

**Rationale:** the cluster meets platform prerequisites; remaining risk is Canal/NetworkPolicy interaction, which is operational rather than a blocker to choosing ambient.

### Make mesh enrollment explicit and reversible

Marketplace namespace or workload enrollment should be controlled through chart values or environment-specific manifests rather than implicit cluster-wide defaults.

**Rationale:** explicit enrollment makes rollback straightforward and avoids surprising local/Tilt behavior. If sidecar mode is selected, namespace injection labels can be managed by GitOps. If ambient mode is selected, enrollment labels can follow the same principle.

### Keep Kafka out of the mesh-enrolled app namespace

Kafka, Kafka Connect, and kafka-ui deploy to a dedicated `kafka` namespace. Marketplace ambient enrollment stays on `ecommerce` only. Debezium reaches CNPG and app secrets in `ecommerce` via cross-namespace DNS (`*.ecommerce.svc`) and Roles/RoleBindings in `appNamespace`.

**Rationale:** namespace-scoped ambient + waypoint on `ecommerce` breaks Strimzi TLS/admin traffic when Kafka shares that namespace. Separating messaging keeps observe-only mesh enrollment simple without per-pod opt-outs.

**Alternatives considered:** label Kafka pods `istio.io/dataplane-mode: none` while leaving them in `ecommerce`. That unblocks mesh quickly but keeps platform data-plane coupled to the app namespace.

### Classify service ports by actual protocol

The marketplace chart should render port names that match the service protocol. The web and payment gateway simulator services are HTTP. Internal gRPC services such as users, products, orders, cart, and payment should expose `grpc` port names.

**Rationale:** Istio uses service metadata for protocol selection and telemetry. Correct names help ensure gRPC requests appear as gRPC rather than generic HTTP or opaque TCP.

### Use VictoriaTraces with Grafana as the telemetry direction

The target observability stack is VictoriaTraces with Grafana, alongside the VictoriaMetrics Kubernetes stack. That prerequisite is implemented as `platform-observability` (`infra/charts/observability`, staging Application `staging-observability` in `monitoring`). Istio observe-mode verification SHOULD use that stack for metrics, logs, traces, and dashboards rather than introducing a temporary tracing UI.

**Rationale:** Istio can emit telemetry into the existing Victoria/Grafana backends. Aligning on the deployed stack avoids building temporary documentation around another tracing tool.

**Dependency note:** `add-victoria-observability-stack` is archived and synced into `openspec/specs/platform-observability`. Grafana/VictoriaTraces checks for Istio are no longer blocked on a missing proposal; they depend on the live staging observability Application.

## Risks / Trade-offs

- **[Istio adds platform complexity]** -> Limit the first change to staging and observe-only behavior.
- **[Protocol naming mistakes reduce telemetry quality]** -> Make protocol an explicit service value and verify rendered Services before rollout.
- **[Sidecar injection can change pod startup/resource behavior]** -> Keep enrollment reversible and verify main browser flows after sync.
- **[Kafka and database traffic may not produce useful L7 telemetry]** -> Treat this change as service request observability first; document lower-layer dependencies separately.
- **[Production rollout may need environment-specific choices]** -> Do not enable production until staging results are known.
- **[Telemetry stack may lag mesh install]** -> Use the deployed `platform-observability` stack; keep Istio verification steps explicit about Grafana/VictoriaTraces access.

## Migration Plan

1. Confirm the staging observability Application (`staging-observability`) is healthy and Grafana/VictoriaTraces are reachable.
2. Add four wrapper charts under `infra/charts/operators/istio/{base,istiod,cni,ztunnel}` pinned to official Istio `1.30.2` with ambient profile on `istiod`/`cni`.
3. Add four staging ArgoCD Applications with sync waves: base → istiod/cni → ztunnel → marketplace enrollment.
4. Sync staging and verify Istio control plane, CNI, and ztunnel health (watch for Canal/NetworkPolicy HBONE port 15008 issues).
5. Add explicit marketplace ambient enrollment for staging.
6. Exercise web, product, cart, checkout, and payment flows.
7. Confirm mesh telemetry shows internal service traffic and expected protocols in Grafana / VictoriaTraces.
8. Document rollback: remove marketplace enrollment first, then disable/remove the Istio Applications if needed.

## Open Questions

- None. Staging ambient readiness and ArgoCD/chart layout are decided above.
