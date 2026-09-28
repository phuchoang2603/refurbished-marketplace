## 1. Argo CD manifests

- [x] 1.1 In `infra/argocd/app-of-apps/templates/applications.tpl`, replace `destination.name` with `server: https://kubernetes.default.svc` for both the generic `apps` loop and the marketplace Application; verify `helm template x infra/argocd/app-of-apps | rg -n 'destination' -A2` shows only the in-cluster server
- [x] 1.2 Add a `syncRetry` block (limit `-1`, backoff `10s` / factor `2` / max `5m`) to `infra/argocd/app-of-apps/values.yaml`, render it under `syncPolicy.retry`, and add `SkipDryRunOnMissingResource=true` to child `syncOptions`; verify the rendered output contains both for all four children
- [x] 1.3 Remove `destinationName` and the management-cluster header comment from `infra/argocd/app-of-apps/values.yaml`; verify `rg destinationName infra/` returns nothing
- [x] 1.4 Replace the destinations in `infra/argocd/project.yaml` with a single `server: https://kubernetes.default.svc`, `namespace: "*"`; verify with `kubectl apply --dry-run=client -f infra/argocd/project.yaml`
- [x] 1.5 In `infra/argocd/{dev,prod}/root.yaml`, drop `destinationName`, rewrite header comments for the per-environment Argo CD and Doppler kubeconfig, and remove the "management cluster" destination comment; verify `rg -n 'management|destinationName|roots/' infra/argocd` returns nothing
- [x] 1.6 Render both roots' effective values through the chart (`helm template x infra/argocd/app-of-apps --set namePrefix=prod ...` with prod value files) and verify four children render with `project: refurbished-marketplace` and in-cluster destinations
- [x] 1.7 Delete the empty untracked `infra/charts/cloudflare-tunnel/` directory; verify it no longer exists

## 2. Deployment docs

- [x] 2.1 Rewrite `docs/deployment/gitops.md` for one Argo CD per environment: fresh install using the Doppler kubeconfig flow from design Decision 4, environment table without "registered Argo cluster"/"platform-dev" rows, ordering section describing retries instead of the readiness gate, per-environment Argo CD UI addresses, and links to talos-proxmox `docs/architecture/gitops.md` and `docs/operations/cluster-access.md`; verify `rg -n 'management|talos-argocd|platform-dev|roots/|apps/README' docs/deployment/gitops.md` is empty
- [x] 2.2 Update `docs/deployment/observability.md` and `docs/deployment/cilium.md` checks and links (drop `talos-argocd.yaml`, point dead talos-proxmox links at current docs); verify the same `rg` is empty for both files
- [x] 2.3 Update `docs/development/secrets.md`: kubeconfig from Doppler `talos-proxmox/<env>`, pre-create `ecommerce`, no management/workload split, state that marketplace does not use the platform `ClusterSecretStore/doppler`; verify `rg -n 'management|workload kubeconfig' docs/development/secrets.md` is empty

## 3. Contributor-facing docs and context

- [x] 3.1 Rewrite the bootstrap block in `docs/development/local-setup.md` (single kubeconfig, no `operators` namespace, no talos-proxmox root apply, no readiness wait) and fix its smoke-check commands; verify `rg -n 'talos-argocd|operators|roots/' docs/development/local-setup.md` is empty
- [x] 3.2 Update the `CONTRIBUTING.md` quick start and prerequisites to the Decision 4 flow; verify `rg -n talos-argocd CONTRIBUTING.md` is empty
- [x] 3.3 Replace management-cluster wording in `README.md` and `docs/architecture.md`; verify `rg -n -i 'management cluster' README.md docs/architecture.md` is empty
- [x] 3.4 Update `openspec/config.yaml` context (GitOps bullets) to describe per-environment Argo CD, the talos-proxmox platform root, and the kubeconfig-from-Doppler flow, and drop Cloudflare Tunnel from this repo's consumers; verify `rg -n 'management|roots/|registered' openspec/config.yaml` is empty

## 4. Verification

- [x] 4.1 Run `rg -n -i 'management cluster|talos-argocd|apps/argocd/roots|platform-dev|platform-prod|destinationName' --glob '!openspec/changes/**'` from the repo root and confirm no matches remain outside archived changes
- [x] 4.2 Run `openspec validate adopt-per-cluster-argocd --strict` and confirm it passes
- [x] 4.3 On dev after talos-proxmox platform bring-up, follow the updated `docs/deployment/gitops.md` fresh-install steps and confirm `kubectl -n argo-cd get applications` shows `dev-root` and its four children Synced/Healthy
