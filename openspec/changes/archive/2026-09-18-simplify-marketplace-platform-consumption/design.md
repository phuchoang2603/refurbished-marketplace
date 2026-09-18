## Context

See proposal.md for motivation and scope. The working tree already removes `infra/charts/operators/*`, `infra/charts/observability`, and their application catalog entries. `talos-proxmox/apps/argocd/roots/{dev,prod}.yaml` defines `platform-dev` and `platform-prod`; its components provide ESO, CNPG, Strimzi, MCK, and Victoria observability. Its MongoDB and Strimzi operators watch workload namespaces.

This repository still has unused observability drift exceptions, a dev root tracking `add-meilisearch-catalog-read-model`, platform installation requirements in current specs, and documentation naming an obsolete platform layout. CI omits MongoDB, app-of-apps, and prod overlays. The platform observability directory has a dashboard template but no marketplace dashboard JSON files.

## Goals / Non-Goals

**Goals:** Make rendered resource ownership explicit, preserve working consumer contracts, and make a fresh installation understandable and verifiable from the remaining charts.

**Non-Goals:** No changes to the sibling repository, live clusters, controller versions, persistence models, or service behavior. Do not add orchestration controllers or compatibility flags. Do not remove service SQL migrations required to initialize empty databases.

## Decisions

### 1. Keep resources with the application that uses them

Retain existing chart boundaries: marketplace, MongoDB, Meilisearch, Kafka, and Cloudflare Tunnel. The operators and their CRDs are platform prerequisites. Keep CNPG clusters, MongoDBCommunity, Kafka/Connect/topics/connectors, ExternalSecrets, VMPodScrape, network policies, and ingress resources. Keep MongoDB namespace-local service accounts/RBAC and Kafka secret-reader RBAC: these support workload operation and do not install operators.

Finish the existing chart deletions and remove observability-only ignore-difference and namespace metadata branches from the application template when unused. Do not delete application custom resources merely because their API group names an operator. Moving databases into the platform repo was considered and rejected because their settings and lifecycle belong to marketplace services.

### 2. Define platform readiness as a prerequisite

Document the platform root deployment, registered destination cluster, healthy controllers and CRDs, default StorageClass, ready `doppler` ClusterSecretStore, and telemetry services as fresh-install prerequisites. Keep application waves as local ordering hints. Do not claim they wait for independent platform roots or guarantee child readiness.

Argo documentation retrieved through Context7 confirms that app-of-apps readiness orchestration requires Application health assessment. The sibling Argo values currently configure only `server.insecure`; this change does not add platform health customization. Adding a second orchestrator or hidden platform-install hook here would blur ownership and is unnecessary for the chosen prerequisite-based deployment model.

Set committed dev and prod roots to `main`; preserve revision inheritance, dev SHA tags, prod `main` tags, destination names, and prod value overlays. Document that required GHCR images must exist before deploying their revision.

### 3. Consume secrets while preserving application setup

Keep ExternalSecrets and current remote-key mappings, including Mongo, Meilisearch, auth, Postgres, and tunnel credentials. Continue documenting manual creation of the environment-specific `operators/doppler-token` and retaining token example manifests because application Doppler setup is still needed. Clarify that ESO and the store are installed/configured in talos-proxmox; this repository only prepares credentials and references the store. No real credentials are written into artifacts or charts.

### 4. Keep application dashboards without restoring the stack

Place marketplace RED/log dashboard JSON and a dedicated ConfigMap template in `infra/charts/refurbished-marketplace`. The marketplace Application owns only those ConfigMaps in the existing `monitoring` namespace. Label them `grafana_dashboard: "1"`, use unique marketplace names, and retain server-side apply. Do not render a monitoring Namespace, Grafana deployment, datasource provisioning, or platform RBAC.

This uses the pinned platform stack's enabled Grafana dashboard sidecar and label selector. Publishing into monitoring avoids adding cross-namespace watch permissions. The bundled Grafana defaults have no folder annotation and disable folders-from-files; therefore folder placement follows platform configuration rather than requiring the old Marketplace folder. Preserve dashboard identity, queries, and correct platform datasource references. Reuse relevant dashboard content from Git history as application assets, without restoring wrapper code or compatibility scaffolding.

Keep VMPodScrape, the scrape network policy, JSON stdout logging, and direct OTLP endpoint `vtsingle-vmks.monitoring.svc.cluster.local:4317`. Verify the dashboard datasource identifiers against platform configuration during implementation. Putting app dashboards entirely in talos-proxmox would couple application metrics/query changes to platform releases and would leave the current missing-dashboard gap unresolved here.

### 5. Keep existing source formatting and remove Helm validation

Remove the Helm CI job and its path filter. Do not add a custom manifest validation script or devenv task. Remove the validation-only kubeconform and yq packages added during this change; retain the existing Helm CLI.

Treefmt-nix provides general YAML integrations (yamlfmt, yamllint, and Prettier), but no built-in Helm rendering or Kubernetes schema-validation integration. Keep the existing Oxfmt integration for plain YAML and its exclusions for Helm templates. No additional formatter is needed.

The previously performed offline manifest checks remain verification evidence for the ownership split, not a maintained automation contract. Source formatting does not establish controller compatibility or runtime readiness.

### 6. Replace current ownership guidance, retain historical records

Rewrite `docs/deployment/gitops.md` and `observability.md`, adjust secrets and CI documentation and any README/diagram ownership references, and update `openspec/config.yaml`. Keep application access, telemetry debugging, secret keys, and empty-database initialization guidance here. Link stack installation/administration to `talos-proxmox/apps/README.md` and its current component paths. Remove transfer/adoption instructions and obsolete platform hook/drift instructions.

Sync the four delta specifications through OpenSpec when finalizing the change. Update the existing argocd-gitops and platform-observability Purpose text at that time so it no longer claims platform installation or historical branch defaults. Archived changes remain historical records; do not sweep them for obsolete paths.

## Risks / Trade-offs

- Platform and application repositories can drift → Document named contracts and compare rendered consumers against the pinned sibling platform configuration; do not silently reinstall a missing platform component.
- Missing prerequisites produce failed reconciliation → Document readiness checks and the platform-first deployment order without inventing cross-root wave guarantees.
- Dashboard assets span namespaces → Own only uniquely named ConfigMaps in monitoring; platform retains namespace and Grafana ownership. Confirm discovery and datasource references during validation.
- Dashboard folder placement changes → Preserve dashboard titles and queries and document platform-controlled placement; no platform change is required.
- Kubeconform skips unknown CRDs → Record this limit and inspect API versions against platform CRDs; runtime smoke checks require a separately authorized deployment.

## Migration Plan

Not applicable: the user chose a fresh start. There are no adoption, preservation, transfer, or legacy rollback steps. Fresh deployment consists of preparing the platform and secrets, confirming required images, then applying the marketplace root. This change prepares repository configuration and documentation only; it does not reset or deploy clusters.
