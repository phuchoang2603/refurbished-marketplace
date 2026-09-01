## ADDED Requirements

### Requirement: Staging Cilium ingress enablement

The staging ArgoCD marketplace Application SHALL be able to enable Cilium edge Gateway API resources through Helm value overlays.

#### Scenario: Staging overlay enables ingress

- **WHEN** staging marketplace values set ingress enablement and host/URL settings
- **THEN** ArgoCD sync renders the Cilium `Gateway` and marketplace `HTTPRoute` resources from the marketplace chart

#### Scenario: Production ingress remains opt-in

- **WHEN** production manifests are rendered before production ingress enablement is chosen
- **THEN** production marketplace workloads do not expose a Cilium ingress Gateway by accident

### Requirement: Argo on Talos owns marketplace

The Talos cluster root Application SHALL enable the marketplace chart via Argo CD. Tilt SHALL NOT be the applier for marketplace Helm. Child Applications SHALL inherit `targetRevision` from the root so branch tracking moves git and (with matching GHCR tags) images together.

#### Scenario: Marketplace is an Argo Application

- **WHEN** the Talos root syncs from Git
- **THEN** a marketplace Application exists and applies `infra/charts/refurbished-marketplace`

#### Scenario: Root revision is inherited

- **WHEN** the root Application `targetRevision` is a branch or `main`
- **THEN** child Applications use that same git revision

### Requirement: Marketplace release owns databases

CNPG Clusters for marketplace services SHALL be resources of the Argo-managed marketplace Helm release. The repository SHALL NOT apply databases out-of-band to protect them from `tilt down`.

#### Scenario: Clusters sync with the chart

- **WHEN** the marketplace Application syncs
- **THEN** CNPG Cluster objects are applied from the chart templates, not from a Tilt `kubectl apply` of `helm template`

## MODIFIED Requirements

### Requirement: App-of-apps per environment

The repository SHALL provide a shared Argo CD app-of-apps Helm chart under `infra/argocd/app-of-apps/` plus a thin root Application for Talos that enables marketplace and sets `global.imageRegistry` / `global.imageTag`. Child Applications SHALL inherit `targetRevision` from the root via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. `infra/argocd/local/` and a Tilt-omitted marketplace Application SHALL NOT exist.

#### Scenario: Talos root application

- **WHEN** the Talos cluster root Application syncs from Git
- **THEN** child Applications exist for operators, `refurbished-marketplace`, and `kafka` as defined

#### Scenario: Talos inherits root revision

- **WHEN** a root Application renders the app-of-apps chart with `targetRevision` parameterized from `$ARGOCD_APP_SOURCE_TARGET_REVISION`
- **THEN** each child Application uses the same Git revision as that root

#### Scenario: Talos shares global image settings

- **WHEN** the Talos root sets `global.imageRegistry` and `global.imageTag`
- **THEN** child Applications that inject global images (for example kafka and marketplace) render those values into their Helm `values`

#### Scenario: Children destine the registered cluster

- **WHEN** a root Application sets `destinationName` to `dev` or `prod`
- **THEN** child Applications destine that Argo CD cluster name (not in-cluster on gpu)

### Requirement: Chart image registry and tag resolution

The `refurbished-marketplace` and `kafka` Helm charts SHALL support `global.imageRegistry` and `global.imageTag`. Cluster deploys SHALL always render GHCR image references (`{registry}/{shortName}:{tag}`). Empty registry / short Colima names SHALL NOT be a supported cluster path. Chart default resources and PVC sizes SHALL match the production-like profile (not a Colima-miniature overlay).

#### Scenario: Remote cluster GHCR reference

- **WHEN** Helm renders with `global.imageRegistry` set and `global.imageTag` set to a git SHA or `main`
- **THEN** a service with `image: web` deploys as `ghcr.io/<repository>/web:<tag>`

#### Scenario: Chart defaults are production-like

- **WHEN** the marketplace chart renders without a Colima overlay
- **THEN** request/limit and database storage defaults are the staging-class sizes, not the 4 CPU / 8 GiB Colima budget

### Requirement: Environment-specific Helm values

The repository SHALL provide a chart-adjacent `values-prod.yaml` overlay (referenced from prod-root via `valueFiles`) for production marketplace hostnames. Dev uses chart `values.yaml`.

#### Scenario: Dev pins the git SHA

- **WHEN** the talos-dev marketplace Application syncs
- **THEN** Helm values set `global.imageRegistry` to the project GHCR path and `global.imageTag` to `$ARGOCD_APP_REVISION`

#### Scenario: Production pulls rolling main tag

- **WHEN** the production marketplace Application syncs
- **THEN** Helm values set `global.imageTag` to `main`

### Requirement: Privileged Pod Security for host-network DaemonSets

Child Applications that deploy host-network or hostPath DaemonSets (Prometheus node-exporter) SHALL set Argo CD `syncPolicy.managedNamespaceMetadata` so `CreateNamespace=true` labels the destination namespace `pod-security.kubernetes.io/enforce=privileged` (and matching audit/warn). Unlabeled namespaces inherit cluster-default PSS baseline (Talos) and those DaemonSets cannot schedule.

#### Scenario: Monitoring namespace allows node-exporter

- **WHEN** the observability Application creates or syncs the `monitoring` namespace
- **THEN** the namespace is labeled for privileged Pod Security so node-exporter can use hostNetwork, hostPID, hostPath, and hostPort

### Requirement: Observability sync ordering

The observability Application SHALL sync before workloads that depend on metrics storage and Grafana.

#### Scenario: Observability precedes application telemetry verification

- **WHEN** sync ordering is evaluated
- **THEN** the observability stack has a sync wave that allows it to become available before checkout trace verification depends on VictoriaTraces

### Requirement: Kafka messaging namespace separation

The staging Kafka Application SHALL deploy Strimzi Kafka, Connect, and UI resources to a dedicated `kafka` namespace so marketplace Cilium L7 or Gateway policies in `ecommerce` do not intercept Kafka TLS traffic.

#### Scenario: Kafka sync targets kafka namespace

- **WHEN** the staging Kafka Application syncs from Git
- **THEN** Kafka cluster resources are applied to the `kafka` namespace rather than `ecommerce`

#### Scenario: Marketplace reaches Kafka across namespaces

- **WHEN** marketplace services publish or consume messages
- **THEN** they use the Kafka bootstrap address in the `kafka` namespace DNS (for example `*.kafka.svc`)

### Requirement: Staging hosted payment URL uses edge route

Staging value overlays SHALL set `HOSTED_PAYMENT_BASE_URL` to the Cloudflare-facing simulator HTTPS base URL when Cilium ingress simulator routing is enabled.

#### Scenario: Staging simulator URL is public edge

- **WHEN** staging ingress with simulator routing is enabled
- **THEN** the web Deployment environment uses the public `https://` simulator hostname, not `http://payment-gateway-simulator:8097` cluster DNS alone and not `http://localhost:8097`

## REMOVED Requirements

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

### Requirement: Staging Istio ingress enablement

The staging ArgoCD marketplace Application SHALL be able to enable Istio edge Gateway API resources through Helm value overlays.

#### Scenario: Staging overlay enables ingress

- **WHEN** staging marketplace values set ingress enablement and host/URL settings
- **THEN** ArgoCD sync renders the Istio `Gateway` and marketplace `HTTPRoute` resources from the marketplace chart

#### Scenario: Production ingress remains opt-in

- **WHEN** production manifests are rendered before production ingress enablement is chosen
- **THEN** production marketplace workloads do not expose an Istio ingress Gateway by accident
