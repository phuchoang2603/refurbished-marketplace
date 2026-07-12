## ADDED Requirements

### Requirement: Staging observability application

The repository SHALL include a staging ArgoCD child Application for the platform observability stack.

#### Scenario: Staging root sync includes observability

- **WHEN** the staging root Application syncs from Git
- **THEN** ArgoCD manages a child Application for the observability stack

#### Scenario: Observability deploys to monitoring namespace

- **WHEN** the staging observability Application syncs
- **THEN** it deploys the observability chart into the `monitoring` namespace

### Requirement: Observability sync ordering

The staging observability Application SHALL sync before workloads or mesh features that depend on metrics storage and Grafana.

#### Scenario: Observability precedes Istio-dependent telemetry

- **WHEN** staging sync ordering is evaluated
- **THEN** the observability stack has a sync wave that allows it to become available before Istio observe-mode dashboard verification depends on it

### Requirement: Observability ArgoCD drift handling

The staging observability Application SHALL include sync and ignore-difference configuration for known `victoria-metrics-k8s-stack` ArgoCD drift sources.

#### Scenario: Generated operator webhook certificates do not cause drift

- **WHEN** ArgoCD compares the VictoriaMetrics operator admission resources
- **THEN** generated validation Secret data and webhook `caBundle` differences are ignored according to the chart guidance

#### Scenario: Generated Grafana password does not cause drift

- **WHEN** ArgoCD compares Grafana resources from the observability stack
- **THEN** generated admin password Secret data and related deployment checksum annotation differences are ignored according to the chart guidance

#### Scenario: Large dashboard ConfigMaps apply successfully

- **WHEN** default dashboard ConfigMaps are applied
- **THEN** the Application or dashboard resources use server-side apply handling so dashboard annotations do not exceed Kubernetes limits

#### Scenario: Pre-delete hooks are not required for closure

- **WHEN** the observability stack is removed by ArgoCD
- **THEN** cleanup does not rely on Helm pre-delete hooks that ArgoCD will ignore

### Requirement: Production observability deferred

The repository SHALL not enable production observability deployment until staging storage, retention, and access patterns are validated.

#### Scenario: Production remains out of scope

- **WHEN** this change is implemented
- **THEN** production ArgoCD observability manifests are not required for closure
