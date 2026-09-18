## Why

Shared operators and observability have moved to `talos-proxmox`, but this repository still contains platform-specific Argo behavior, stale deployment guidance, and specifications that claim ownership of those components. A fresh deployment needs a clear platform prerequisite contract while retaining the custom resources and supporting configuration that make the marketplace work.

## What Changes

- **BREAKING**: Complete removal of operator and platform observability installation from this repository; require the corresponding `talos-proxmox` platform deployment before deploying marketplace workloads.
- Keep application CNPG clusters, MongoDBCommunity and workload RBAC, Strimzi resources, ExternalSecrets, VMPodScrape, Cilium policies, and Gateway resources.
- Remove unused platform drift handling and historical branch defaults from Argo configuration; use `main` as the committed default for both roots while retaining dev SHA image tags.
- Keep marketplace RED/log dashboards as application-owned ConfigMaps consumed by platform Grafana, without restoring the observability wrapper.
- Rewrite deployment and secret documentation, project context, and affected specifications around fresh installation and platform consumption.
- Remove Helm CI validation without adding a custom local script or devenv task; retain existing treefmt/Oxfmt formatting for plain YAML.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `argocd-gitops`: Marketplace-only ownership, platform prerequisites, fresh deployment defaults, and accurate application ordering semantics.
- `external-secrets`: Consume the platform-owned ESO and Doppler store while retaining application secret mappings and token preparation guidance.
- `platform-observability`: Consume platform telemetry services and provision only marketplace dashboards and scrape resources here.
- `github-actions-ci`: Remove Helm validation and retain existing local source formatting.

## Impact

Changes affect `infra/argocd`, remaining `infra/charts`, `.github/workflows/ci.yml`, deployment/development docs, and `openspec/config.yaml`. Existing working-tree operator/observability deletions are part of the intended final state. `talos-proxmox` is a read-only reference and external prerequisite, not an implementation target for this change.

Non-goals: data migration, resource adoption or ownership transfer, legacy compatibility, live cluster reset/deployment, platform operator upgrades, service business-logic changes, and rewriting archived change history. Keep database schema initialization needed for empty databases.
