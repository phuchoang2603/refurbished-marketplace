## Context

See proposal.md for motivation. Relevant state:

- `talos-proxmox`'s platform root installs Argo CD (`argo-cd` namespace, release `argo-cd`) per environment, plus a bootstrap chart that creates the `talos-proxmox` AppProject and `platform` root. Argo CD runs with server-side apply and server-side diff enabled globally.
- Platform child Applications use `destination.server: https://kubernetes.default.svc`, unlimited `retry` with backoff (`10s`, factor `2`, max `5m`), and `SkipDryRunOnMissingResource=true`.
- The environment kubeconfig lives in Doppler at `talos-proxmox/<dev|prod>` key `KUBECONFIG`. The per-environment Argo CD UIs are `http://10.69.11.254` (dev) and `http://10.69.12.254` (prod).
- This repo's app-of-apps renders children with `destination.name: {{ destinationName }}`; `project.yaml` lists `name: dev` / `name: prod` destinations; both roots set `destinationName`.
- `SecretStore/doppler` is rendered in `ecommerce` by three charts and reads `ecommerce/doppler-token` (Doppler project `refurbished-marketplace`, configs `dev` / `prd`). That stays.

## Goals / Non-Goals

**Goals:**

- Marketplace roots sync on each environment's own Argo CD with no cluster registration.
- A fresh environment needs exactly three operator steps after platform bring-up: fetch kubeconfig, apply token Secret, apply project + root.
- Child convergence behavior matches the platform chart so ordering against platform components no longer needs a human check.

**Non-Goals:**

- Sharing templates or values with the talos-proxmox platform chart (no cross-repo Helm dependency).
- Automating the root install (see proposal non-goals).

## Decisions

### 1. Hard-code the in-cluster destination server

Children use `destination.server: https://kubernetes.default.svc`; `destinationName` is deleted from `values.yaml`, the template, and both roots.

- Alternative: keep a `destination` value defaulting to in-cluster. Rejected: no remaining topology needs another target, and the repo convention is to remove legacy knobs rather than keep compatibility paths.

### 2. Mirror the platform's retry and dry-run policy in app-of-apps

Add a `syncRetry` block to `infra/argocd/app-of-apps/values.yaml` (same shape and defaults as the platform chart) and render it under `syncPolicy.retry` for every child, plus `SkipDryRunOnMissingResource=true` in `syncOptions`. The root Applications keep their current policy; they only create child Application objects, whose CRD already exists.

- Alternative: keep the manual "verify platform readiness" gate. Rejected: platform children already converge asynchronously via the same mechanism, and the gate was the main source of multi-step fresh-install docs.
- Alternative: rely on Argo CD Application health assessment for cross-root ordering. Rejected: not configured on the platform Argo CD and would require talos-proxmox changes.

### 3. AppProject destinations are in-cluster only

`infra/argocd/project.yaml` keeps its name, namespace, source repo, and cluster-resource whitelist, and replaces the three destinations with a single `server: https://kubernetes.default.svc`, `namespace: "*"`. This covers the root (`argo-cd`) and children (`ecommerce`, `kafka`, `monitoring` for dashboards).

### 4. Operator workflow: one kubeconfig per environment

Docs standardize on:

```bash
export CLUSTER_ENV=dev   # or prod
doppler secrets get KUBECONFIG --plain --project talos-proxmox --config "$CLUSTER_ENV" \
  > "$HOME/.kube/talos-${CLUSTER_ENV}.yaml"
export KUBECONFIG="$HOME/.kube/talos-${CLUSTER_ENV}.yaml"
kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.<dev|prd>.secret.yaml
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply --server-side -f infra/argocd/<dev|prod>/root.yaml
```

The `ecommerce` namespace is pre-created because the token Secret must exist before the root's `CreateNamespace=true` would create it; the stale `operators` namespace step is dropped. Explicit `--project talos-proxmox` is required because `devenv.nix` defaults `DOPPLER_PROJECT` to `refurbished-marketplace`. The file path `~/.kube/talos-<env>.yaml` matches talos-proxmox's cluster-access guide, so existing `--kubeconfig` examples keep working; only `talos-argocd.yaml` references are removed.

### 5. Documentation boundaries

- `docs/deployment/gitops.md` owns the per-environment Argo CD model, fresh install, and checks; links to talos-proxmox's `docs/architecture/gitops.md` and `docs/operations/cluster-access.md` replace the dead `apps/README.md` link.
- `docs/development/secrets.md` owns the token bootstrap and states marketplace does not use the platform `ClusterSecretStore/doppler`.
- `docs/development/local-setup.md`, `CONTRIBUTING.md`, `README.md`, `docs/architecture.md`, `docs/deployment/{cilium,observability}.md` get targeted wording/command fixes only.
- `openspec/config.yaml` context is updated so future artifacts stop describing the management cluster and the removed talos-proxmox roots.

## Risks / Trade-offs

- [Existing Applications on a surviving management Argo CD would still point at `destination.name`] → The management cluster is already torn down upstream; no migration path is kept. If one still exists, delete its marketplace roots before applying the new ones to avoid two controllers owning the same resources.
- [Unlimited retries hide a genuinely broken child] → Backoff caps at 5 minutes; failures stay visible as `OperationState` errors in the environment's Argo CD UI, same as platform components.
- [`SkipDryRunOnMissingResource` lets a typo'd CRD kind reach sync] → Surfaced by the sync failure and retried; charts are unchanged by this change, so no new kinds are introduced.
- [Token Secret remains manual] → Accepted per decision; documented as the single manual prerequisite.

## Migration Plan

1. Merge to `main`.
2. For each environment after talos-proxmox platform bring-up: run the Decision 4 steps.
3. Confirm `kubectl -n argo-cd get applications` shows `dev-root`/`prod-root` and four children Synced/Healthy alongside `platform`.

Rollback: revert the commit; there is no management cluster to roll back to, so rollback only restores the previous (non-functional) manifests.
