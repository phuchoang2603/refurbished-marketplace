## RENAMED Requirements

- FROM: `### Requirement: Argo on management cluster destines Talos workload clusters`
- TO: `### Requirement: Per-environment Argo CD deploys marketplace locally`

## REMOVED Requirements

### Requirement: App-of-apps per environment

**Reason**: Its contract set `destinationName` on management-cluster roots so children targeted registered clusters; that topology no longer exists.

**Migration**: Replaced by "Local app-of-apps per environment": roots run in each environment's Argo CD and children destine `https://kubernetes.default.svc`.

## MODIFIED Requirements

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

### Requirement: Cloudflare Tunnel is platform-owned

The marketplace repository SHALL NOT include an Argo CD child Application or Helm chart for `cloudflared`. The platform repository SHALL deploy the connector in the `cloudflare-tunnel` namespace, with its token delivered by the platform's External Secrets `ClusterSecretStore`.

#### Scenario: Marketplace root omits cloudflared

- **WHEN** `dev-root` or `prod-root` syncs from this repository
- **THEN** no Cloudflare Tunnel child Application is created

#### Scenario: Platform root provides cloudflared

- **WHEN** the matching talos-proxmox platform Application syncs
- **THEN** it manages the tunnel deployment using the platform-delivered token Secret

## ADDED Requirements

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
