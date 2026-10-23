# ArgoCD GitOps

## Purpose

Define marketplace GitOps delivery through the prod root on the cluster's Argo CD, consuming shared platform operators and observability from `talos-proxmox` while deploying application resources to the local `prod` cluster used as the development server.

## Requirements

### Requirement: Istio ingress enablement on dest clusters

The marketplace Application SHALL render Istio Gateway API resources when chart ingress is enabled. Chart defaults SHALL enable ingress for `shop-dev`; production SHALL use `shop` via `values-prod.yaml`.

#### Scenario: Chart defaults enable ingress

- **WHEN** marketplace values use chart defaults in a Helm render
- **THEN** the output includes the Istio Gateway and marketplace HTTPRoutes without requiring a cluster deployment

#### Scenario: Production overlay enables prod hosts

- **WHEN** prod-root applies `values-prod.yaml`
- **THEN** production marketplace workloads expose an Istio ingress Gateway for `shop`

### Requirement: Namespace enrollment has one GitOps owner

The marketplace root SHALL declaratively create and enroll `ecommerce` and `kafka` in Istio ambient without competing namespace metadata owners, out-of-band labeling, or automatic enrollment of shared platform namespaces.

#### Scenario: Applications sync from empty state

- **WHEN** the production root starts against a fresh platform
- **THEN** both application namespaces have the ambient enrollment label before strict workload traffic depends on it

### Requirement: Argo CD deploys marketplace locally

The prod cluster's Argo CD (namespace `argo-cd`, installed by the talos-proxmox platform root) SHALL host the `prod-root` Application and enable the marketplace chart. Children SHALL destine the local cluster through `https://kubernetes.default.svc`. The repository SHALL NOT depend on a management cluster or on Argo CD cluster registrations. Child Applications SHALL inherit `targetRevision` from the root so branch tracking moves git and (with matching GHCR tags) images together. The committed prod root SHALL track `main` for both its source and child Applications, with immutable SHA image tags pinned in the production overlays.

#### Scenario: Marketplace is an Argo Application

- **WHEN** the Talos root syncs from Git
- **THEN** a marketplace Application exists and applies `infra/charts/refurbished-marketplace`

#### Scenario: Root revision is inherited

- **WHEN** the root Application `targetRevision` is a branch or `main`
- **THEN** child Applications use that same git revision

#### Scenario: Root runs in the target environment

- **WHEN** an operator installs `prod-root`
- **THEN** the root Application is created in `argo-cd` on that same environment's cluster, and no other cluster's Argo CD is involved

### Requirement: Marketplace release owns databases

CNPG Clusters for marketplace services that still use Postgres SHALL be resources of the Argo-managed marketplace Helm release. The marketplace release SHALL NOT deploy a `products-db` Cluster after catalog cutover. The repository SHALL NOT apply remaining databases out-of-band to protect them from `tilt down`.

#### Scenario: Clusters sync with the chart

- **WHEN** the marketplace Application syncs
- **THEN** CNPG Cluster objects for remaining Postgres services are applied from the chart templates and no products Cluster is rendered

### Requirement: Chart image registry and tag resolution

The `refurbished-marketplace` and `kafka` Helm charts SHALL support `global.imageRegistry` and `global.imageTag`. Cluster deploys SHALL always render GHCR image references (`{registry}/{shortName}:{tag}`). Chart default resources and PVC sizes SHALL match the Talos profile.

#### Scenario: Remote cluster GHCR reference

- **WHEN** Helm renders with `global.imageRegistry` set and `global.imageTag` set to a git SHA or `main`
- **THEN** a service with `image: web` deploys as `ghcr.io/<repository>/web:<tag>`

#### Scenario: Chart defaults match Talos

- **WHEN** the marketplace chart renders with default values
- **THEN** request/limit and database storage defaults are the Talos sizes

### Requirement: Production Helm values

The repository SHALL provide a chart-adjacent `values-prod.yaml` overlay (referenced from prod-root via `valueFiles`) for production marketplace hostnames. Chart `values.yaml` defaults SHALL remain available for standalone rendering.

#### Scenario: Production pulls pinned images

- **WHEN** the production marketplace Application syncs
- **THEN** Helm values set `global.imageTag` to the immutable SHA pinned in `values-prod.yaml`

### Requirement: Loose sync ordering

Child Applications SHALL retain local ordering hints for data stores, marketplace, and Kafka. Waves SHALL NOT be described as ordering the platform root against the marketplace root. Deployment guidance SHALL NOT require a manual platform-readiness check before the marketplace root is applied; convergence against platform operators, CRDs, storage, and telemetry SHALL rely on child sync retries. The `ecommerce/doppler-token` Secret SHALL remain a manual prerequisite for application secrets.

#### Scenario: Operator wave before apps

- **WHEN** the talos-proxmox platform root has installed Argo CD and the marketplace root is applied while platform child Applications are still converging
- **THEN** marketplace child Applications keep retrying and reach Synced once the platform operators and CRDs they consume are available

#### Scenario: Child readiness

- **WHEN** deployment guidance describes child sync waves
- **THEN** it states that waves order submission only, and that waiting for child Application health is not guaranteed by the current Argo CD configuration

### Requirement: GitOps documentation

The repository SHALL document the Argo CD layout (the prod cluster's Argo CD hosting `prod-root` and destining its local cluster), `values-prod.yaml` overlays, immutable image tags pinned in the production overlays, and fresh-install prerequisites (talos-proxmox platform bring-up for prod, the prod kubeconfig fetched from the `talos-proxmox` Doppler project, the ecommerce Doppler token, and Cloudflare origin configuration). It SHALL identify the talos-proxmox platform root and `apps/components` as platform references and SHALL NOT reference a management cluster, `platform-dev` / `platform-prod` roots, or `apps/argocd/roots/`. It SHALL NOT prescribe legacy ownership transfer or data migration.

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a Talos deploy
- **THEN** documentation explains app-of-apps paths, value overlays, and pinned SHA tags

#### Scenario: Contributor uses the prod kubeconfig

- **WHEN** a contributor follows the deploy guide for prod
- **THEN** they fetch the prod `KUBECONFIG` from Doppler and use it for the token Secret, the AppProject, and the root, with no separate management kubeconfig

### Requirement: Kafka messaging namespace separation

The Kafka Application SHALL deploy Strimzi Kafka, Connect, and UI resources to a dedicated ambient-enrolled `kafka` namespace rather than `ecommerce`. Kafka's broker-level TLS behavior SHALL remain independent of Istio transport encryption, and Kafka Connect SHALL reach MongoDB in the enrolled `ecommerce` namespace.

#### Scenario: Kafka sync targets kafka namespace

- **WHEN** the Kafka Application syncs from Git
- **THEN** Kafka cluster resources are applied in an ambient-enrolled `kafka` namespace rather than `ecommerce`

#### Scenario: Marketplace reaches Kafka across namespaces

- **WHEN** marketplace services publish or consume messages
- **THEN** they use the Kafka bootstrap address in the `kafka` namespace DNS (for example `*.kafka.svc`)

#### Scenario: Connector reaches catalog database

- **WHEN** Kafka Connect synchronizes MongoDB changes
- **THEN** its cross-namespace database connection works under strict ambient mTLS

### Requirement: Cloudflare Tunnel is platform-owned

The marketplace repository SHALL NOT include an Argo CD child Application or Helm chart for `cloudflared`. The platform repository SHALL deploy the connector in the `cloudflare-tunnel` namespace, with its token delivered by the platform's External Secrets `ClusterSecretStore`.

#### Scenario: Marketplace root omits cloudflared

- **WHEN** `prod-root` syncs from this repository
- **THEN** no Cloudflare Tunnel child Application is created

#### Scenario: Platform root provides cloudflared

- **WHEN** the matching talos-proxmox platform Application syncs
- **THEN** it manages the tunnel deployment using the platform-delivered token Secret

### Requirement: MongoDB Community Application in ecommerce

The repository SHALL include an Argo CD child Application that applies the MongoDB Community replica set (and related secrets/policy owned by that chart) into the `ecommerce` namespace. That Application SHALL NOT template an `ecommerce` Namespace object (the marketplace Application already destines that namespace).

#### Scenario: Database CR destines ecommerce

- **WHEN** the Mongo database Application syncs
- **THEN** the MongoDB Community custom resource is applied in `ecommerce`

#### Scenario: Prod may overlay size

- **WHEN** prod-root applies Mongo chart value overlays
- **THEN** production MAY raise replica-set members or resources without changing the Community operator path

### Requirement: GitOps docs include Mongo

GitOps documentation SHALL identify the platform-owned MCK operator and marketplace-owned Mongo database Application, their namespaces, and MongoDB as the catalog system of record.

#### Scenario: Contributor finds Mongo in the deploy table

- **WHEN** a contributor reads the deploy guide
- **THEN** it distinguishes the external operator in operators from MongoDBCommunity and workload RBAC in ecommerce, and identifies products as a Mongo consumer

### Requirement: Inventory workload in marketplace chart

The marketplace Helm release SHALL deploy an `inventory` Service and Deployment in `ecommerce` with GHCR image tags matching other marketplace services.

#### Scenario: Inventory syncs with the chart

- **WHEN** the marketplace Application syncs
- **THEN** an inventory Deployment and Service exist in `ecommerce`

### Requirement: Product creation event transport

The GitOps configuration SHALL retain the `products.created` topic and SHALL deploy a Debezium MongoDB outbox connector against the catalog outbox collection, using Mongo credentials and the same EventRouter identity/key/payload/tracing mapping as other outbox connectors. It SHALL NOT deploy a Postgres products-outbox CDC connector. Inventory SHALL continue to subscribe to creation events with its own consumer group.

#### Scenario: Creation transport syncs

- **WHEN** the Kafka and marketplace Applications sync
- **THEN** the products creation topic, Mongo outbox connector, and inventory subscription SHALL be configured without a Postgres `products-outbox` connector and without replacing the inventory reservation event transport

### Requirement: Meilisearch Application in ecommerce

The repository SHALL include an Argo CD child Application that deploys Meilisearch into the `ecommerce` namespace. That Application SHALL NOT template an `ecommerce` Namespace object. The deployment SHALL require platform External Secrets readiness and retain the local data-store wave before marketplace workloads; this wave SHALL NOT imply ordering against the independent platform root.

#### Scenario: Search engine destines ecommerce

- **WHEN** the Meilisearch Application syncs
- **THEN** the Meilisearch workload is applied in `ecommerce`

#### Scenario: Search service can reference Meilisearch

- **WHEN** the marketplace Application syncs after Meilisearch is Healthy
- **THEN** the search workload can be configured with the in-cluster Meilisearch URL without a second search cluster

### Requirement: Search workload in marketplace chart

The marketplace Helm release SHALL deploy a `search` Service and Deployment in `ecommerce` with GHCR image tags matching other marketplace services.

#### Scenario: Search syncs with the chart

- **WHEN** the marketplace Application syncs
- **THEN** a search Deployment and Service exist in `ecommerce`

### Requirement: GitOps docs include Meilisearch

GitOps documentation SHALL list the Meilisearch Application, the search marketplace workload, namespaces, and local sync-wave hints relative to marketplace and the external platform operator prerequisite.

#### Scenario: Contributor finds Meilisearch in the deploy table

- **WHEN** a contributor reads the GitOps deploy guide
- **THEN** documentation names the Meilisearch Application, the search service, `ecommerce`, and that Meilisearch is a catalog projection rather than listing SoT

### Requirement: Marketplace consumes platform custom resources

Marketplace deployment SHALL retain CNPG Cluster, MongoDBCommunity, Kafka, KafkaNodePool, KafkaConnect, KafkaTopic, KafkaConnector, ExternalSecret, SecretStore, Gateway, HTTPRoute, PeerAuthentication, and DestinationRule resources needed by enabled workloads. It SHALL retain MongoDB workload service accounts/RBAC, Kafka connector secret-access RBAC, and schema initialization jobs. It SHALL NOT render CiliumNetworkPolicy or install shared operators, their CRDs, the CNI, or the platform observability stack.

#### Scenario: Consumer resources survive cleanup

- **WHEN** default and production application charts are rendered
- **THEN** enabled workload custom resources and supporting RBAC remain, no Cilium policy is rendered, and no shared operator installation or platform stack is rendered

#### Scenario: Empty databases

- **WHEN** marketplace workloads start against new empty databases
- **THEN** schema initialization still runs before dependent service operation

### Requirement: Local app-of-apps

The repository SHALL provide a shared Argo CD app-of-apps Helm chart under `infra/argocd/app-of-apps/` plus a thin `prod-root` Application that enables marketplace and sets `global.imageRegistry` / `global.imageTag`. The chart and roots SHALL NOT expose a destination cluster name. Child Applications SHALL inherit `targetRevision` from the root via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Child Applications SHALL retry failed syncs without a retry limit, using backoff, and SHALL skip dry-run for resources whose CRDs are not yet installed.

#### Scenario: Talos root application

- **WHEN** the Talos cluster root Application syncs from Git
- **THEN** child Applications exist for `refurbished-marketplace`, MongoDB, Meilisearch, and Kafka, with no operator, Cloudflare Tunnel, or platform observability Application

#### Scenario: Talos inherits root revision

- **WHEN** a root Application renders the app-of-apps chart with `targetRevision` parameterized from `$ARGOCD_APP_SOURCE_TARGET_REVISION`
- **THEN** each child Application uses the same Git revision as that root

#### Scenario: Talos shares global image settings

- **WHEN** the Talos root sets `global.imageRegistry` and `global.imageTag`
- **THEN** child Applications that inject global images (for example kafka and marketplace) render those values into their Helm `values`

#### Scenario: Children destine the local cluster

- **WHEN** the app-of-apps chart renders child Applications for prod
- **THEN** every child destination server is `https://kubernetes.default.svc` and no child references a destination cluster name

#### Scenario: Children wait out missing platform CRDs

- **WHEN** a child Application syncs before a platform operator has installed its CRDs
- **THEN** the sync fails without a hard dry-run error and is retried with backoff until the CRDs exist

### Requirement: Marketplace AppProject scope

The repository SHALL provide a `refurbished-marketplace` AppProject in `argo-cd` whose source is this repository and whose only destination server is `https://kubernetes.default.svc`. It SHALL NOT list named remote clusters.

#### Scenario: Project admits local children

- **WHEN** the AppProject is installed and the root syncs
- **THEN** the root and every child Application are admitted by the project

#### Scenario: Project rejects remote destinations

- **WHEN** an Application in the project targets a destination other than `https://kubernetes.default.svc`
- **THEN** Argo CD rejects it as outside the project's destinations

### Requirement: Kafka requests reflect observed usage

Kafka broker and Kafka Connect CPU requests SHALL be sized to their observed steady-state usage rather than a nominal minimum, so the scheduler sees their real load.

#### Scenario: Broker and Connect requests

- **WHEN** the kafka chart renders with default values
- **THEN** broker and Connect containers request at least 250m CPU

### Requirement: Checkout payment gateway configuration per environment

The production overlay SHALL configure Checkout with the in-cluster payment gateway Edge address in the `payment-gateway` namespace, the marketplace's gateway integration ID, the gateway API key and webhook signing secret from the Checkout gateway Secret, and the buyer return URL on the `shop` host. Chart defaults SHALL leave the gateway disabled, rendering no gateway address or Secret reference. The chart SHALL NOT render a Payment service or payment database.

#### Scenario: Production render

- **WHEN** the production marketplace Application renders
- **THEN** the Checkout Deployment has the gateway address and Secret reference, and no `payment` Deployment, Service or CNPG Cluster exists

#### Scenario: Default render

- **WHEN** the marketplace chart renders with default values
- **THEN** Checkout has no gateway address or Secret configuration
