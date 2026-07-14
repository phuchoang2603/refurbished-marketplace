# ArgoCD GitOps

## Purpose

Define local Colima (`infra/argocd/local/`) and staging (`infra/argocd/staging/`) GitOps delivery via Argo CD app-of-apps, chart-adjacent Helm overlays, and coordinated GHCR image tags. Production app-of-apps is deferred until added.

## Requirements

### Requirement: App-of-apps per environment

The repository SHALL provide ArgoCD Application manifests under `infra/argocd/<environment>/` for `local` and `staging` (and `production` when added). Staging SHALL include a root Application that syncs child Applications for operators, the `refurbished-marketplace` Helm chart, and the `kafka` Helm chart. Local Argo SHALL sync operators, Istio, Kafka, observability, and Cloudflare Tunnel, and SHALL NOT manage the `refurbished-marketplace` Application (Tilt owns that chart locally).

#### Scenario: Staging root application

- **WHEN** the staging cluster root Application syncs from Git
- **THEN** child Applications exist for operators, `refurbished-marketplace`, and `kafka` in the `ecommerce` and `operators` namespaces as defined

#### Scenario: Local Argo omits marketplace

- **WHEN** a developer uses local Argo CD (`infra/argocd/local/`) on Colima
- **THEN** no marketplace Application is present; chart default `values.yaml` is applied by Tilt for the marketplace release

#### Scenario: Local infra uses chart defaults

- **WHEN** local Argo syncs infra Applications
- **THEN** chart default `values.yaml` is used; staging overlays live in chart-adjacent `values-staging.yaml` files

### Requirement: Environment-specific Helm values

The repository SHALL provide Helm value overlays as chart-adjacent `values-staging.yaml` files (referenced from staging Applications via `valueFiles`) for marketplace, Istio CNI, and observability where needed. Staging overlays SHALL set `global.imageTag` to `main` for marketplace/kafka images. Production overlays SHALL set `global.imageTag` to a commit SHA for coordinated releases when production is added.

#### Scenario: Staging pulls rolling main tag

- **WHEN** the staging marketplace Application syncs
- **THEN** Helm values set `global.imageRegistry` to the project GHCR path and `global.imageTag` to `main`

#### Scenario: Production pins commit SHA

- **WHEN** the production marketplace Application syncs after a promotion
- **THEN** Helm values set `global.imageTag` to the promoted commit SHA shared by all service images

### Requirement: Chart image registry and tag resolution

The `refurbished-marketplace` and `kafka` Helm charts SHALL support `global.imageRegistry` and `global.imageTag`. When `global.imageRegistry` is empty, templates SHALL render service image fields unchanged for local Colima builds. When set, templates SHALL render images as `{registry}/{shortName}:{tag}`.

#### Scenario: Local image names unchanged

- **WHEN** Helm renders with default chart values and empty `global.imageRegistry`
- **THEN** service deployments reference short names such as `web`

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

The repository SHALL document the Argo CD layout (local vs staging), chart-adjacent `values-staging.yaml` overlays, image tag promotion for production when added, and prerequisites that remain outside Git (Argo bootstrap, Doppler token, ClusterSecretStore, Cloudflare tunnel token).

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a local or staging deploy
- **THEN** documentation explains app-of-apps paths, value overlays, local Tilt + Argo DX, and SHA promotion for remote clusters

### Requirement: Observability application

The repository SHALL include local and staging Argo CD child Applications for the platform observability stack.

#### Scenario: Root sync includes observability

- **WHEN** the local or staging root Application syncs from Git
- **THEN** Argo CD manages a child Application for the observability stack

#### Scenario: Observability deploys to monitoring namespace

- **WHEN** an observability Application syncs
- **THEN** it deploys the observability chart into the `monitoring` namespace

### Requirement: Observability sync ordering

The observability Application SHALL sync before workloads or mesh features that depend on metrics storage and Grafana.

#### Scenario: Observability precedes Istio-dependent telemetry

- **WHEN** sync ordering is evaluated
- **THEN** the observability stack has a sync wave that allows it to become available before Istio observe-mode dashboard verification depends on it

### Requirement: Observability ArgoCD drift handling

Local and staging observability Applications SHALL include sync and ignore-difference configuration for known `victoria-metrics-k8s-stack` ArgoCD drift sources.

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

### Requirement: Staging Istio ingress enablement

The staging ArgoCD marketplace Application SHALL be able to enable Istio edge Gateway API resources through Helm value overlays.

#### Scenario: Staging overlay enables ingress

- **WHEN** staging marketplace values set ingress enablement and host/URL settings
- **THEN** ArgoCD sync renders the Istio `Gateway` and marketplace `HTTPRoute` resources from the marketplace chart

#### Scenario: Production ingress remains opt-in

- **WHEN** production manifests are rendered before production ingress enablement is chosen
- **THEN** production marketplace workloads do not expose an Istio ingress Gateway by accident

### Requirement: Staging hosted payment URL uses edge route

Staging value overlays SHALL set `HOSTED_PAYMENT_BASE_URL` to the Cloudflare-facing simulator HTTPS base URL when Istio ingress simulator routing is enabled.

#### Scenario: Staging simulator URL is public edge

- **WHEN** staging ingress with simulator routing is enabled
- **THEN** the web Deployment environment uses the public `https://` simulator hostname, not `http://payment-gateway-simulator:8097` cluster DNS alone and not `http://localhost:8097`

### Requirement: Cloudflare Tunnel application

The repository SHALL include Argo CD child Applications under local and staging that deploy in-cluster `cloudflared` for the marketplace edge.

#### Scenario: Root sync includes cloudflare-tunnel

- **WHEN** the local or staging root Application syncs from Git
- **THEN** Argo CD manages a child Application for the Cloudflare Tunnel connector in the `cloudflare-tunnel` namespace

#### Scenario: Tunnel token comes from External Secrets

- **WHEN** the cloudflare-tunnel chart syncs with External Secrets enabled
- **THEN** the tunnel token Secret is populated from Doppler via an ExternalSecret rather than committed to Git
