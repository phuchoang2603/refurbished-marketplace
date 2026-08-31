## ADDED Requirements

### Requirement: Staging Cilium ingress enablement

The staging ArgoCD marketplace Application SHALL be able to enable Cilium edge Gateway API resources through Helm value overlays.

#### Scenario: Staging overlay enables ingress

- **WHEN** staging marketplace values set ingress enablement and host/URL settings
- **THEN** ArgoCD sync renders the Cilium `Gateway` and marketplace `HTTPRoute` resources from the marketplace chart

#### Scenario: Production ingress remains opt-in

- **WHEN** production manifests are rendered before production ingress enablement is chosen
- **THEN** production marketplace workloads do not expose a Cilium ingress Gateway by accident

## MODIFIED Requirements

### Requirement: Environment-specific Helm values

The repository SHALL provide Helm value overlays as chart-adjacent `values-staging.yaml` files (referenced from staging Applications via `valueFiles`) for marketplace and observability where needed. Staging overlays SHALL set `global.imageTag` to `main` for marketplace/kafka images. Production overlays SHALL set `global.imageTag` to a commit SHA for coordinated releases when production is added.

#### Scenario: Staging pulls rolling main tag

- **WHEN** the staging marketplace Application syncs
- **THEN** Helm values set `global.imageRegistry` to the project GHCR path and `global.imageTag` to `main`

#### Scenario: Production pins commit SHA

- **WHEN** the production marketplace Application syncs after a promotion
- **THEN** Helm values set `global.imageTag` to the promoted commit SHA shared by all service images

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
