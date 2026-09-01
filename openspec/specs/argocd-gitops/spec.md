# ArgoCD GitOps

## Purpose

Define GitOps delivery via a shared Argo CD app-of-apps Helm chart (`infra/argocd/app-of-apps`). Argo CD runs on the gpu cluster. Thin `dev-root` and `prod-root` Applications destine registered clusters `dev` and `prod`. Chart defaults are shop-dev; production hosts live in `values-prod.yaml`. `dev-root` `targetRevision` MAY stay on a feature branch until operators retarget it.

## Requirements

### Requirement: Cilium ingress enablement on dest clusters

The marketplace Application SHALL render Cilium edge Gateway API resources when the chart has ingress enabled. Chart defaults SHALL enable ingress for `shop-dev` / `pay-dev`. Production SHALL enable ingress for `shop` / `pay` via `values-prod.yaml`.

#### Scenario: Dev chart defaults enable ingress

- **WHEN** talos-dev marketplace values use chart defaults
- **THEN** Argo CD sync renders the Cilium `Gateway` and marketplace `HTTPRoute` resources

#### Scenario: Production overlay enables prod hosts

- **WHEN** prod-root applies `values-prod.yaml`
- **THEN** production marketplace workloads expose a Cilium ingress Gateway for `shop` / `pay`

### Requirement: Argo on gpu destines Talos workload clusters

Root Applications on the gpu cluster SHALL enable the marketplace chart via Argo CD. Children SHALL destine registered clusters `dev` or `prod`. Tilt SHALL NOT be the applier for marketplace Helm. Child Applications SHALL inherit `targetRevision` from the root so branch tracking moves git and (with matching GHCR tags) images together. `dev-root` `targetRevision` MAY remain on a feature branch after merge until operators retarget it.

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

### Requirement: App-of-apps per environment

The repository SHALL provide a shared Argo CD app-of-apps Helm chart under `infra/argocd/app-of-apps/` plus thin `dev-root` and `prod-root` Applications on gpu that enable marketplace and set `global.imageRegistry` / `global.imageTag` and `destinationName`. Child Applications SHALL inherit `targetRevision` from the root via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. `infra/argocd/local/` and a Tilt-omitted marketplace Application SHALL NOT exist.

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

### Requirement: Payment gateway simulator in marketplace chart

The repository SHALL deploy `payment-gateway-simulator` from the `refurbished-marketplace` Helm chart.

#### Scenario: Simulator enabled

- **WHEN** the marketplace chart syncs
- **THEN** a `payment-gateway-simulator` Deployment and Service exist in `ecommerce`

### Requirement: Loose sync ordering

Child ArgoCD Applications SHALL use sync waves so operators sync before the marketplace chart and the marketplace chart syncs before the kafka chart.

#### Scenario: Operator wave before apps

- **WHEN** a full environment sync runs
- **THEN** operator Applications have a lower sync wave than marketplace and kafka Applications

### Requirement: GitOps documentation

The repository SHALL document the Argo CD layout (gpu roots destining `dev` / `prod`), `values-prod.yaml` overlays, image tags (`$ARGOCD_APP_REVISION` vs `:main`), and prerequisites that remain outside Git (Argo bootstrap, Doppler token, ClusterSecretStore, Cloudflare Public Hostname origin DNS).

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a Talos deploy
- **THEN** documentation explains app-of-apps paths, value overlays, and SHA vs `:main` tags

### Requirement: Observability application

The repository SHALL include Argo CD child Applications for the platform observability stack.

#### Scenario: Root sync includes observability

- **WHEN** `dev-root` or `prod-root` syncs from Git
- **THEN** Argo CD manages a child Application for the observability stack

#### Scenario: Observability deploys to monitoring namespace

- **WHEN** an observability Application syncs
- **THEN** it deploys the observability chart into the `monitoring` namespace

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

### Requirement: Observability ArgoCD drift handling

Observability Applications SHALL include sync and ignore-difference configuration for known `victoria-metrics-k8s-stack` ArgoCD drift sources.

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

### Requirement: Kafka messaging namespace separation

The Kafka Application SHALL deploy Strimzi Kafka, Connect, and UI resources to a dedicated `kafka` namespace so marketplace Gateway policies in `ecommerce` do not intercept Kafka TLS traffic.

#### Scenario: Kafka sync targets kafka namespace

- **WHEN** the Kafka Application syncs from Git
- **THEN** Kafka cluster resources are applied to the `kafka` namespace rather than `ecommerce`

#### Scenario: Marketplace reaches Kafka across namespaces

- **WHEN** marketplace services publish or consume messages
- **THEN** they use the Kafka bootstrap address in the `kafka` namespace DNS (for example `*.kafka.svc`)

### Requirement: Hosted payment URL uses edge route

Chart defaults SHALL set `HOSTED_PAYMENT_BASE_URL` to the Cloudflare-facing `pay-dev` HTTPS URL. Production overlay SHALL use `pay`.

#### Scenario: Simulator URL is public edge

- **WHEN** ingress with simulator routing is enabled
- **THEN** the web Deployment environment uses the public `https://` simulator hostname, not `http://payment-gateway-simulator:8097` cluster DNS alone and not `http://localhost:8097`

### Requirement: Cloudflare Tunnel application

The repository SHALL include Argo CD child Applications that deploy in-cluster `cloudflared` for the marketplace edge.

#### Scenario: Root sync includes cloudflare-tunnel

- **WHEN** `dev-root` or `prod-root` syncs from Git
- **THEN** Argo CD manages a child Application for the Cloudflare Tunnel connector in the `cloudflare-tunnel` namespace

#### Scenario: Tunnel token comes from External Secrets

- **WHEN** the cloudflare-tunnel chart syncs with External Secrets enabled
- **THEN** the tunnel token Secret is populated from Doppler via an ExternalSecret rather than committed to Git
