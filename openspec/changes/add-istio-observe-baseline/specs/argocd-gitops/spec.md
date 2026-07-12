## ADDED Requirements

### Requirement: GitOps-managed Istio baseline

The repository SHALL provide GitOps-managed configuration for installing the Istio platform baseline in staging before marketplace workloads depend on mesh enrollment.

#### Scenario: Staging sync installs Istio

- **WHEN** the staging root Application syncs from Git
- **THEN** ArgoCD manages four Istio Applications backed by wrapper charts under `infra/charts/operators/istio/{base,istiod,cni,ztunnel}` for observe-only marketplace mesh enrollment

#### Scenario: Istio wrappers pin official charts

- **WHEN** the Istio operator wrapper charts are built
- **THEN** each wrapper depends on the matching official Istio Helm chart pinned to version `1.30.2`, with ambient profile enabled for `istiod` and `cni`

#### Scenario: Istio syncs before enrolled workloads

- **WHEN** a full staging environment sync runs
- **THEN** Istio platform resources are ordered `base` → `istiod`/`cni` → `ztunnel` before marketplace workloads that require mesh enrollment

### Requirement: Environment-scoped mesh rollout

The repository SHALL scope the first Istio rollout to staging unless production mesh enablement is explicitly configured.

#### Scenario: Staging has mesh enrollment configuration

- **WHEN** staging marketplace values or manifests are applied
- **THEN** marketplace workloads can be enrolled in Istio through GitOps-managed configuration

#### Scenario: Production is not implicitly enrolled

- **WHEN** production manifests are rendered before production mesh enablement is chosen
- **THEN** production marketplace workloads are not enrolled in Istio by accident

#### Scenario: Production waits for staging validation

- **WHEN** staging mesh enrollment has not been verified successfully
- **THEN** production Istio installation and marketplace enrollment remain out of scope for the first rollout
