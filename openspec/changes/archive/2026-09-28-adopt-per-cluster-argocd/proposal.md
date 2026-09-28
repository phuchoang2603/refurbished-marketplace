## Why

`talos-proxmox` removed the management cluster: dev and prod each now run their own Argo CD in `argo-cd`, installed by the OpenTofu `platform` root, with no remote-cluster registrations. This repository's roots, AppProject, and child Applications still target registered clusters `dev` / `prod` from a management Argo CD, so they can no longer sync, and the docs point at deleted `talos-proxmox` roots and a `talos-argocd.yaml` kubeconfig that no longer exists.

## What Changes

- **BREAKING**: Marketplace roots (`dev-root`, `prod-root`) are applied to the Argo CD running inside the target environment, not to a management cluster.
- **BREAKING**: Child Applications destine `https://kubernetes.default.svc`; the `destinationName` value is removed from the app-of-apps chart and roots.
- The `refurbished-marketplace` AppProject allows only in-cluster destinations.
- Child Applications adopt the platform's convergence settings (unlimited sync retry with backoff and `SkipDryRunOnMissingResource`), so the marketplace root can be applied while platform components are still converging.
- The manual platform-readiness gate before applying the marketplace root is replaced by retry-based convergence; the `ecommerce/doppler-token` Secret remains the one manual prerequisite.
- Operator workflow fetches the environment kubeconfig from Doppler (`talos-proxmox` project, `dev` / `prod` config) and uses that single kubeconfig for the token Secret, the AppProject, and the root.
- Docs, root comments, `CONTRIBUTING.md`, README, and OpenSpec project context drop the management cluster, `platform-dev` / `platform-prod` roots, `../talos-proxmox/apps/argocd/roots/*.yaml`, and dead `talos-proxmox` links.
- Cloudflare Tunnel wording reflects that the platform delivers its token through its own ESO `ClusterSecretStore`.
- Remove the empty, untracked `infra/charts/cloudflare-tunnel/` leftover.

## Non-goals

- Installing the marketplace root from `talos-proxmox` (bootstrap chart or platform root) or from a new OpenTofu root in this repository.
- Moving marketplace secrets to the platform `ClusterSecretStore/doppler` or having `talos-proxmox` provision `ecommerce/doppler-token`. The namespaced `SecretStore/doppler` and hand-applied token stay.
- Renaming the Doppler `prd` config to `prod`.
- Restructuring this repository's docs into architecture/operations/reference sections.
- Changing chart workloads, image tagging, sync waves, or value overlays.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `argocd-gitops`: Roots run on each environment's own Argo CD; children destine the local cluster; sync ordering relies on retries rather than a manual platform-readiness gate; documentation describes per-environment Argo CD and Doppler-sourced kubeconfigs; Cloudflare Tunnel token delivery wording is updated.
- `external-secrets`: Bootstrap token is applied with the environment kubeconfig from Doppler (no management/workload split); stale `operators` namespace scenario and Cloudflare "does not use ESO" statement are corrected.

## Impact

- `infra/argocd/app-of-apps/{values.yaml,templates/applications.tpl}`, `infra/argocd/project.yaml`, `infra/argocd/{dev,prod}/root.yaml`
- `docs/deployment/{gitops.md,cilium.md,observability.md}`, `docs/development/{local-setup.md,secrets.md}`, `docs/architecture.md`, `README.md`, `CONTRIBUTING.md`
- `openspec/config.yaml` project context; `openspec/specs/{argocd-gitops,external-secrets}` via deltas
- Operators: existing dev/prod clusters must receive the AppProject and root through their own Argo CD after `talos-proxmox` platform bring-up.
