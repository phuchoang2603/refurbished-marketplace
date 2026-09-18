## MODIFIED Requirements

### Requirement: Argo on management cluster destines Talos workload clusters

Root Applications on the management cluster SHALL enable the marketplace chart via Argo CD. Children SHALL destine registered clusters `dev` or `prod`. Child Applications SHALL inherit `targetRevision` from the root so branch tracking moves git and (with matching GHCR tags) images together. Committed dev and prod roots SHALL default to `main`; dev SHALL retain SHA-based images and prod SHALL retain the rolling `main` image tag.

#### Scenario: Marketplace is an Argo Application

- **WHEN** the Talos root syncs from Git
- **THEN** a marketplace Application exists and applies `infra/charts/refurbished-marketplace`

#### Scenario: Root revision is inherited

- **WHEN** the root Application `targetRevision` is a branch or `main`
- **THEN** child Applications use that same git revision

### Requirement: App-of-apps per environment

The repository SHALL provide a shared Argo CD app-of-apps Helm chart under `infra/argocd/app-of-apps/` plus thin `dev-root` and `prod-root` Applications on management cluster that enable marketplace and set `global.imageRegistry` / `global.imageTag` and `destinationName`. Child Applications SHALL inherit `targetRevision` from the root via `$ARGOCD_APP_SOURCE_TARGET_REVISION`.

#### Scenario: Talos root application

- **WHEN** the Talos cluster root Application syncs from Git
- **THEN** child Applications exist for `refurbished-marketplace`, MongoDB, Meilisearch, Kafka, and Cloudflare Tunnel, with no operator or platform observability Application

#### Scenario: Talos inherits root revision

- **WHEN** a root Application renders the app-of-apps chart with `targetRevision` parameterized from `$ARGOCD_APP_SOURCE_TARGET_REVISION`
- **THEN** each child Application uses the same Git revision as that root

#### Scenario: Talos shares global image settings

- **WHEN** the Talos root sets `global.imageRegistry` and `global.imageTag`
- **THEN** child Applications that inject global images (for example kafka and marketplace) render those values into their Helm `values`

#### Scenario: Children destine the registered cluster

- **WHEN** a root Application sets `destinationName` to `dev` or `prod`
- **THEN** child Applications destine that Argo CD cluster name (not in-cluster on management cluster)

### Requirement: Loose sync ordering

Child Applications SHALL retain local ordering hints for data stores, marketplace, Kafka, and Cloudflare Tunnel. Deployment guidance SHALL distinguish these hints from readiness guarantees and SHALL require platform readiness before marketplace sync; waves SHALL NOT be described as ordering independent platform and marketplace roots.

#### Scenario: Operator wave before apps

- **WHEN** a fresh environment is prepared
- **THEN** the platform root is applied and its operators, CRDs, storage, secret store, and telemetry services are verified before the marketplace root is applied

#### Scenario: Child readiness

- **WHEN** deployment guidance describes child sync waves
- **THEN** it states that waiting for child Application health requires Argo Application health assessment and is not guaranteed by the current platform configuration

### Requirement: GitOps documentation

The repository SHALL document the Argo CD layout (management cluster roots destining `dev` / `prod`), `values-prod.yaml` overlays, image tags (`$ARGOCD_APP_REVISION` vs `:main`), and fresh-install prerequisites (Argo cluster registration, the talos-proxmox platform root, ready CRDs/operators, default storage, Doppler token and ready ClusterSecretStore, telemetry services, and Cloudflare origin configuration). It SHALL identify platform-dev/platform-prod and apps/components as platform references and SHALL NOT prescribe legacy ownership transfer or data migration.

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a Talos deploy
- **THEN** documentation explains app-of-apps paths, value overlays, and SHA vs `:main` tags

### Requirement: GitOps docs include Mongo

GitOps documentation SHALL identify the platform-owned MCK operator and marketplace-owned Mongo database Application, their namespaces, and MongoDB as the catalog system of record.

#### Scenario: Contributor finds Mongo in the deploy table

- **WHEN** a contributor reads the deploy guide
- **THEN** it distinguishes the external operator in operators from MongoDBCommunity and workload RBAC in ecommerce, and identifies products as a Mongo consumer

### Requirement: Meilisearch Application in ecommerce

The repository SHALL include an Argo CD child Application that deploys Meilisearch into the `ecommerce` namespace. That Application SHALL NOT template an `ecommerce` Namespace object. The deployment SHALL require platform External Secrets readiness and retain the local data-store wave before marketplace workloads; this wave SHALL NOT imply ordering against the independent platform root.

#### Scenario: Search engine destines ecommerce

- **WHEN** the Meilisearch Application syncs
- **THEN** the Meilisearch workload is applied in `ecommerce`

#### Scenario: Search service can reference Meilisearch

- **WHEN** the marketplace Application syncs after Meilisearch is Healthy
- **THEN** the search workload can be configured with the in-cluster Meilisearch URL without a second search cluster

### Requirement: GitOps docs include Meilisearch

GitOps documentation SHALL list the Meilisearch Application, the search marketplace workload, namespaces, and local sync-wave hints relative to marketplace and the external platform operator prerequisite.

#### Scenario: Contributor finds Meilisearch in the deploy table

- **WHEN** a contributor reads the GitOps deploy guide
- **THEN** documentation names the Meilisearch Application, the search service, `ecommerce`, and that Meilisearch is a catalog projection rather than listing SoT

## ADDED Requirements

### Requirement: Marketplace consumes platform custom resources

Marketplace deployment SHALL retain CNPG Cluster, MongoDBCommunity, Kafka, KafkaNodePool, KafkaConnect, KafkaTopic, KafkaConnector, ExternalSecret, VMPodScrape, CiliumNetworkPolicy, Gateway, and HTTPRoute resources needed by enabled workloads. It SHALL retain MongoDB workload service accounts/RBAC, Kafka connector secret-access RBAC, and schema initialization jobs. It SHALL NOT install shared operators, their CRDs, the Doppler ClusterSecretStore, or the platform observability stack.

#### Scenario: Consumer resources survive cleanup

- **WHEN** dev and prod application charts are rendered
- **THEN** enabled workload custom resources and supporting RBAC remain, and no shared operator installation or platform stack is rendered

#### Scenario: Empty databases

- **WHEN** marketplace workloads start against new empty databases
- **THEN** schema initialization still runs before dependent service operation

## REMOVED Requirements

### Requirement: Observability application

**Reason**: This is platform installation behavior owned by talos-proxmox.

**Migration**: None. Fresh installations consume the platform provided by talos-proxmox; existing resource adoption is outside scope.

### Requirement: Privileged Pod Security for host-network DaemonSets

**Reason**: This is platform installation behavior owned by talos-proxmox.

**Migration**: None. Fresh installations consume the platform provided by talos-proxmox; existing resource adoption is outside scope.

### Requirement: Observability sync ordering

**Reason**: This is platform installation behavior owned by talos-proxmox.

**Migration**: None. Fresh installations consume the platform provided by talos-proxmox; existing resource adoption is outside scope.

### Requirement: Observability ArgoCD drift handling

**Reason**: This is platform installation behavior owned by talos-proxmox.

**Migration**: None. Fresh installations consume the platform provided by talos-proxmox; existing resource adoption is outside scope.

### Requirement: MCK operator Application

**Reason**: This is platform installation behavior owned by talos-proxmox.

**Migration**: None. Fresh installations consume the platform provided by talos-proxmox; existing resource adoption is outside scope.
