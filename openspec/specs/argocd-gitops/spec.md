# ArgoCD GitOps

## Purpose

Define marketplace GitOps delivery through Argo CD roots on each environment's own Argo CD, consuming shared platform operators and observability from `talos-proxmox` while deploying application resources to the local `dev` or `prod` cluster.

## Requirements

### Requirement: Cilium ingress enablement on dest clusters

The marketplace Application SHALL render Cilium edge Gateway API resources when the chart has ingress enabled. Chart defaults SHALL enable ingress for `shop-dev` / `pay-dev`. Production SHALL enable ingress for `shop` / `pay` via `values-prod.yaml`.

#### Scenario: Dev chart defaults enable ingress

- **WHEN** talos-dev marketplace values use chart defaults
- **THEN** Argo CD sync renders the Cilium `Gateway` and marketplace `HTTPRoute` resources

#### Scenario: Production overlay enables prod hosts

- **WHEN** prod-root applies `values-prod.yaml`
- **THEN** production marketplace workloads expose a Cilium ingress Gateway for `shop` / `pay`

### Requirement: Per-environment Argo CD deploys marketplace locally

Each environment's own Argo CD (namespace `argo-cd`, installed by the talos-proxmox platform root) SHALL host that environment's marketplace root Application and enable the marketplace chart. Children SHALL destine the local cluster through `https://kubernetes.default.svc`. The repository SHALL NOT depend on a management cluster or on Argo CD cluster registrations. Child Applications SHALL inherit `targetRevision` from the root so branch tracking moves git and (with matching GHCR tags) images together. Committed dev and prod roots SHALL default to `main`; dev SHALL retain SHA-based images and prod SHALL retain the rolling `main` image tag.

#### Scenario: Marketplace is an Argo Application

- **WHEN** the Talos root syncs from Git
- **THEN** a marketplace Application exists and applies `infra/charts/refurbished-marketplace`

#### Scenario: Root revision is inherited

- **WHEN** the root Application `targetRevision` is a branch or `main`
- **THEN** child Applications use that same git revision

#### Scenario: Root runs in the target environment

- **WHEN** an operator installs `dev-root` or `prod-root`
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

Child Applications SHALL retain local ordering hints for data stores, marketplace, and Kafka. Waves SHALL NOT be described as ordering the platform root against the marketplace root. Deployment guidance SHALL NOT require a manual platform-readiness check before the marketplace root is applied; convergence against platform operators, CRDs, storage, and telemetry SHALL rely on child sync retries. The `ecommerce/doppler-token` Secret SHALL remain a manual prerequisite for application secrets.

#### Scenario: Operator wave before apps

- **WHEN** the talos-proxmox platform root has installed Argo CD and the marketplace root is applied while platform child Applications are still converging
- **THEN** marketplace child Applications keep retrying and reach Synced once the platform operators and CRDs they consume are available

#### Scenario: Child readiness

- **WHEN** deployment guidance describes child sync waves
- **THEN** it states that waves order submission only, and that waiting for child Application health is not guaranteed by the current Argo CD configuration

### Requirement: GitOps documentation

The repository SHALL document the Argo CD layout (one Argo CD per environment, each hosting its own marketplace root and destining its local cluster), `values-prod.yaml` overlays, image tags (`$ARGOCD_APP_REVISION` vs `:main`), and fresh-install prerequisites (talos-proxmox platform bring-up for that environment, the environment kubeconfig fetched from the `talos-proxmox` Doppler project, the ecommerce Doppler token, and Cloudflare origin configuration). It SHALL identify the talos-proxmox platform root and `apps/components` as platform references and SHALL NOT reference a management cluster, `platform-dev` / `platform-prod` roots, or `apps/argocd/roots/`. It SHALL NOT prescribe legacy ownership transfer or data migration.

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a Talos deploy
- **THEN** documentation explains app-of-apps paths, value overlays, and SHA vs `:main` tags

#### Scenario: Contributor uses one kubeconfig per environment

- **WHEN** a contributor follows the deploy guide for dev or prod
- **THEN** they fetch that environment's `KUBECONFIG` from Doppler and use it for the token Secret, the AppProject, and the root, with no separate management kubeconfig

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

### Requirement: Cloudflare Tunnel is platform-owned

The marketplace repository SHALL NOT include an Argo CD child Application or Helm chart for `cloudflared`. The platform repository SHALL deploy the connector in the `cloudflare-tunnel` namespace, with its token delivered by the platform's External Secrets `ClusterSecretStore`.

#### Scenario: Marketplace root omits cloudflared

- **WHEN** `dev-root` or `prod-root` syncs from this repository
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

Marketplace deployment SHALL retain CNPG Cluster, MongoDBCommunity, Kafka, KafkaNodePool, KafkaConnect, KafkaTopic, KafkaConnector, ExternalSecret, SecretStore, VMPodScrape, CiliumNetworkPolicy, Gateway, and HTTPRoute resources needed by enabled workloads. It SHALL retain MongoDB workload service accounts/RBAC, Kafka connector secret-access RBAC, and schema initialization jobs. It SHALL NOT install shared operators, their CRDs, or the platform observability stack.

#### Scenario: Consumer resources survive cleanup

- **WHEN** dev and prod application charts are rendered
- **THEN** enabled workload custom resources and supporting RBAC remain, and no shared operator installation or platform stack is rendered

#### Scenario: Empty databases

- **WHEN** marketplace workloads start against new empty databases
- **THEN** schema initialization still runs before dependent service operation

### Requirement: Local app-of-apps per environment

The repository SHALL provide a shared Argo CD app-of-apps Helm chart under `infra/argocd/app-of-apps/` plus thin `dev-root` and `prod-root` Applications that enable marketplace and set `global.imageRegistry` / `global.imageTag`. The chart and roots SHALL NOT expose a destination cluster name. Child Applications SHALL inherit `targetRevision` from the root via `$ARGOCD_APP_SOURCE_TARGET_REVISION`. Child Applications SHALL retry failed syncs without a retry limit, using backoff, and SHALL skip dry-run for resources whose CRDs are not yet installed.

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

- **WHEN** the app-of-apps chart renders child Applications for dev or prod
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
