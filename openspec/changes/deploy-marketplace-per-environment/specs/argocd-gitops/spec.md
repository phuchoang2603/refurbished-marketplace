## MODIFIED Requirements

### Requirement: Argo CD deploys marketplace locally

Each environment SHALL host its own marketplace root and children on its local Argo CD. Each root SHALL watch its environment deployment ref; that release commit SHALL pin child Git revisions and immutable GHCR image tags to the successfully published source release. Dev SHALL receive trusted PRs and prod main releases, without cross-cluster registration.

#### Scenario: Marketplace is an Argo Application

- **WHEN** the Talos root syncs from Git
- **THEN** a marketplace Application exists and applies `infra/charts/refurbished-marketplace`

#### Scenario: Root revision is inherited

- **WHEN** the root Application `targetRevision` is an immutable release SHA
- **THEN** child Applications use that same immutable git revision

#### Scenario: Root runs in the target environment

- **WHEN** an operator installs `prod-root`
- **THEN** the root Application is created in `argo-cd` on that same environment's cluster, and no other cluster's Argo CD is involved

### Requirement: Production Helm values

Production SHALL use chart-adjacent values-prod.yaml overlays for hostnames and storage sizing. CI SHALL provide the immutable image tag in the coordinated deployment ref to marketplace and Kafka after all images publish.

#### Scenario: Production pulls pinned images

- **WHEN** the production marketplace Application syncs
- **THEN** the environment deployment ref sets `global.imageTag` to the immutable published tag

### Requirement: Loose sync ordering

Children SHALL retain ordering hints and retry against platform CRDs and operators. The operator SHALL bootstrap the environment-scoped ecommerce/doppler-token and root once; subsequent releases SHALL reconcile from Git without CI access to Kubernetes.

#### Scenario: Operator wave before apps

- **WHEN** the talos-proxmox platform root has installed Argo CD and the marketplace root is applied while platform child Applications are still converging
- **THEN** marketplace child Applications keep retrying and reach Synced once the platform operators and CRDs they consume are available

#### Scenario: Child readiness

- **WHEN** deployment guidance describes child sync waves
- **THEN** it states that waves order submission only, and that waiting for child Application health is not guaranteed by the current Argo CD configuration

### Requirement: GitOps documentation

Deployment documentation SHALL describe independent dev/prod roots, matching Git/image revision pins, local secret bootstrap and the platform-owned dependencies. It SHALL NOT describe prod as the development server or require a LAN LoadBalancer.

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a Talos deploy
- **THEN** documentation explains app-of-apps paths, value overlays, and pinned SHA tags

#### Scenario: Contributor uses the prod kubeconfig

- **WHEN** a contributor follows the deploy guide for prod
- **THEN** they fetch the matching environment `KUBECONFIG` from Doppler and use it for the token Secret, the AppProject, and the root, with no separate management kubeconfig
