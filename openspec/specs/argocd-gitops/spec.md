# ArgoCD GitOps

## Purpose

Define staging and production GitOps delivery via ArgoCD app-of-apps, environment Helm overlays, and coordinated GHCR image tags without requiring Tilt on remote clusters.

## Requirements

### Requirement: App-of-apps per environment

The repository SHALL provide ArgoCD Application manifests under `infra/argocd/<environment>/` for `staging` and `production` only. Each environment SHALL include a root Application that syncs child Applications for operators, the `refurbished-marketplace` Helm chart, and the `kafka` Helm chart.

#### Scenario: Staging root application

- **WHEN** the staging cluster root Application syncs from Git
- **THEN** child Applications exist for operators, `refurbished-marketplace`, and `kafka` in the `ecommerce` and `operators` namespaces as defined

#### Scenario: No dev ArgoCD overlay

- **WHEN** a developer uses Tilt for local development
- **THEN** no `infra/argocd/values/dev/` or dev Application set is required; chart default `values.yaml` remains the dev source

### Requirement: Environment-specific Helm values

The repository SHALL provide Helm value overlays at `infra/argocd/values/staging/` and `infra/argocd/values/production/` for the marketplace and kafka charts. Staging overlays SHALL set `global.imageTag` to `main`. Production overlays SHALL set `global.imageTag` to a commit SHA for coordinated releases.

#### Scenario: Staging pulls rolling main tag

- **WHEN** the staging marketplace Application syncs
- **THEN** Helm values set `global.imageRegistry` to the project GHCR path and `global.imageTag` to `main`

#### Scenario: Production pins commit SHA

- **WHEN** the production marketplace Application syncs after a promotion
- **THEN** Helm values set `global.imageTag` to the promoted commit SHA shared by all service images

### Requirement: Chart image registry and tag resolution

The `refurbished-marketplace` and `kafka` Helm charts SHALL support `global.imageRegistry` and `global.imageTag`. When `global.imageRegistry` is empty, templates SHALL render service image fields unchanged for local Tilt. When set, templates SHALL render images as `{registry}/{shortName}:{tag}`.

#### Scenario: Tilt local image names unchanged

- **WHEN** Helm renders with default chart values and empty `global.imageRegistry`
- **THEN** service deployments reference short names such as `refurbished-marketplace/web`

#### Scenario: Remote cluster GHCR reference

- **WHEN** Helm renders with `global.imageRegistry` set and `global.imageTag` set to `main`
- **THEN** a service with `image: web` deploys as `ghcr.io/<repository>/web:main`

### Requirement: Payment gateway simulator in marketplace chart

The repository SHALL deploy `payment-gateway-simulator` from the `refurbished-marketplace` Helm chart. Staging and production value overlays SHALL configure `HOSTED_PAYMENT_BASE_URL` to reach the in-cluster simulator service.

#### Scenario: Simulator enabled in staging

- **WHEN** the staging marketplace chart syncs
- **THEN** a `payment-gateway-simulator` Deployment and Service exist in `ecommerce`

#### Scenario: Web uses in-cluster simulator URL remotely

- **WHEN** staging or production marketplace values are applied
- **THEN** the web service `HOSTED_PAYMENT_BASE_URL` targets the in-cluster simulator, not `localhost`

### Requirement: Loose sync ordering

Child ArgoCD Applications SHALL use sync waves so operators sync before the marketplace chart and the marketplace chart syncs before the kafka chart.

#### Scenario: Operator wave before apps

- **WHEN** a full environment sync runs
- **THEN** operator Applications have a lower sync wave than marketplace and kafka Applications

### Requirement: GitOps documentation

The repository SHALL document the ArgoCD layout, staging vs production value locations, image tag promotion for production, and prerequisites that remain outside Git (Argo bootstrap, Doppler token, ClusterSecretStore).

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a staging or production deploy
- **THEN** development documentation explains app-of-apps paths, value overlays, and SHA promotion without requiring Tilt

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

### Requirement: Kafka messaging namespace separation

The staging Kafka Application SHALL deploy Strimzi Kafka, Connect, and UI resources to a dedicated `kafka` namespace so marketplace ambient enrollment in `ecommerce` does not intercept Kafka TLS traffic.

#### Scenario: Kafka sync targets kafka namespace

- **WHEN** the staging Kafka Application syncs from Git
- **THEN** Kafka cluster resources are applied to the `kafka` namespace rather than `ecommerce`

#### Scenario: Marketplace reaches Kafka across namespaces

- **WHEN** marketplace services publish or consume messages
- **THEN** they use the Kafka bootstrap address in the `kafka` namespace DNS (for example `*.kafka.svc`)
