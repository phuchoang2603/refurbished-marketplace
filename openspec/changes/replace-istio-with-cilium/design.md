## Context

Today: Tilt on Colima owns marketplace Helm + `docker_build` + templ/tailwind watches; Argo `local-root` skips marketplace; staging Argo tracks `main` and GHCR `:main`. Talos already runs Cilium. Tilt against Talos would need a registry and `live_update` because Talos does not share the laptop Docker daemon.

New intent: **no Tilt deploy path.** One Talos cluster, Argo CD is the only applier, git branch + GHCR SHA is the inner loop.

templ `_templ.go` files are already committed; Tailwind CSS is copied from the repo in `web.Dockerfile`. Watches were convenience, not a cluster requirement.

Follow-on: `add-cilium-mesh-policy-and-canary`.

## Goals / Non-Goals

**Goals:**

- One dataplane: Cilium CNI + Cilium Gateway API + Hubble L4 on Talos.
- Same browser contract: Cloudflare HTTPS, HTTP origin, host-based routes, forwarded proto/host headers.
- Dev and prod share one cluster shape: same Argo apps, GHCR images, observability, CNPG-in-chart, shop/pay Gateway. Overlays only for revision, image SHA, and secrets if required.
- CI publishes images Argo can pull: SHA tags on `main` and on PRs; prune PR-only GHCR versions when the PR closes.
- Remove Istio; kafka stays out of L7 intercept.

**Non-Goals:**

- Tilt, Colima/k3s, Tiltfile workarounds, and a second “local” Helm personality.
- Preview environments per PR (extra namespaces, hosts, CNPG clusters). One live `ecommerce` on the cluster.
- Hubble L7 metrics / Istio-RED replacement.
- Cilium as an Argo Application.
- Production app-of-apps (still deferred).
- Mesh/Gateway proxy spans.

## Decisions

### 1. Cilium is cluster-owned in talos-proxmox; do not copy Helm values here

Source of truth: sibling repo **talos-proxmox** `apps/values/cilium.yaml` (installed by `apps/bootstrap.sh`, chart 1.18.13). Live talos-dev `helm get values cilium` already includes WireGuard, Envoy L7, `cluster.name/id`, Gateway API, and Hubble. L2 pool and Hubble/Longhorn/Argo Gateways live in `apps/manifests/env/*/network.yaml` + `routes.yaml`.

This repo does **not** keep a second Cilium values file (it would drift). Marketplace GitOps only consumes `gatewayClassName: cilium`. Follow-on `add-cilium-mesh-policy-and-canary` may add policies/routes; it must not helm-upgrade Cilium.

### 2. Argo CD is the only deploy path; delete Tilt-era quirks

One Talos cluster, one app-of-apps root, marketplace always an Argo Application, images always GHCR. Chart **defaults** are that cluster (prod-like), not a Colima profile that staging overlays back to reality.

**Delete (legacy / Tilt-only):**

| Quirk                                                                                                             | Why it existed                                   | Replacement                                                                |
| ----------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ | -------------------------------------------------------------------------- |
| `Tiltfile` (install Argo, Gateway API CRDs, Doppler secret, `local-root`, `docker_build`, watches, port-forwards) | Colima inner loop                                | Git push → CI → Argo; `kubectl` port-forward if needed                     |
| Strip Namespace from Helm + Tilt `ecommerce-namespace` labels                                                     | Tilt prune/recreate deleted CNPG                 | Argo `CreateNamespace` + `managedNamespaceMetadata` only (no Istio labels) |
| Out-of-band `helm template databases.tpl \| kubectl apply`                                                        | `tilt down` deleted CNPG PVCs / Debezium offsets | Marketplace chart owns CNPG Clusters; `tilt down` is gone                  |
| `local-root` with `marketplace.enabled: false`                                                                    | Tilt owned the app chart                         | Single root; marketplace on                                                |
| Empty `global.imageRegistry` → short image names                                                                  | Tilt loaded images into Colima Docker            | Always `ghcr.io/.../name:sha`                                              |
| Marketplace `values.yaml` Colima CPU/mem / 512Mi PVCs                                                             | Fit 4 CPU / 8 GiB k3s                            | Defaults = current staging-class requests/limits/storage                   |
| Observability chart defaults apps-only (no node-exporter/ksm/Alertmanager/default dashboards)                     | Colima RAM                                       | Defaults = full platform stack (today’s `values-staging.yaml` shape)       |
| Dual Istio CNI `platform: k3s` vs RKE2 overlays                                                                   | Colima vs old staging                            | Istio charts deleted                                                       |
| `mesh.tpl` comments + ambient labels for Tilt vs Argo namespace fight                                             | Two appliers                                     | No waypoint; Argo owns ns metadata                                         |
| devenv `DOCKER_HOST` Colima k8s socket, `tilt` package                                                            | Local k8s + Tilt                                 | Docker only if Testcontainers needs it; drop Tilt                          |
| Tilt-applied `doppler-token.dev.secret.yaml`                                                                      | Local bootstrap                                  | Same pattern as remote: bootstrap Secret once, ESO `ClusterSecretStore`    |
| Tilt extra Gateway API CRD apply                                                                                  | Colima had no CRDs                               | Talos/Cilium already has Gateway API                                       |
| Docs/PR template/`CONTRIBUTING` `tilt up` on Colima                                                               | Old DX                                           | Argo + GHCR                                                                |
| `.dev` Cloudflare hosts as chart default                                                                          | Second Colima front door                         | Cluster uses the real shop/pay hostnames; no second ingress profile        |

**Allowed to differ between “what’s live on main” and a branch:** `targetRevision`, `global.imageTag` (SHA), Doppler config if secrets must not mix. Not a second cluster shape.

**Rationale:** matching dev and prod means the same YAML path, not a miniature stack that only resembles prod after overlays.

**Alternatives considered:** keep Colima-sized defaults with a `values-talos.yaml` overlay (still two personalities); Tilt `live_update` (rejected).

### 3. Branch tracking is one pointer, not N preview envs

Default: root `targetRevision: main`, `global.imageTag: main` (or the SHA `:main` points at).

To run a feature on the cluster: point that **same** root at the branch and set `global.imageTag` to the PR head SHA CI pushed. One `ecommerce` namespace — only one git revision live at a time.

Do **not** ApplicationSet-per-PR in this change (would need `pr-N` namespaces, extra shop hosts, extra CNPG).

**Rationale:** homelab has one shop. Branch tracking is “what is live,” not “every PR gets a URL.”

**Alternatives considered:** ApplicationSet PR generator (correct for multi-preview, too much here); only deploy after merge (simplest; then PR images are optional until you want to flip the pointer).

### 4. CI builds PR images; delete PR-only tags after close — yes, with rules

**Build on PR (ideal if you want to Argo-track a branch before merge):**

- Reuse the release matrix (or path-filter it).
- Push immutable `ghcr.io/<repo>/<image>:<git-sha>` (`github.sha` of the PR head).
- Optional moving tag `pr-<n>` for humans — Argo should pin **SHA**, not `pr-N` (that tag moves on every push).
- Do not overwrite `:main` from a PR.

**After merge/close, delete PR-only versions:**

- Trigger `pull_request: types: [closed]`.
- Delete GHCR package versions whose only useful tags are `pr-<n>` or SHAs that are **not** on `main` and not currently specified in Argo.
- **Do not** delete `:main`, production pins, or a SHA Argo is still rolling out.
- Squash merge ⇒ PR SHA ≠ merge SHA ⇒ safe to delete the PR SHA images after close.
- Merge commit ⇒ same: PR head SHA is usually unused after merge; still delete `pr-*`.

This is good hygiene (GHCR storage, fewer stale tags). It is **not** a substitute for pinning Argo to SHAs. Deleting too early while the cluster still runs `imageTag: <pr-sha>` will 404 ImagePullBackOff — always retarget Argo to `main` before or as part of close.

**Cheaper alternative if PRs are not deployed:** skip PR pushes; only `main` builds images. Branch tracking then cannot deploy app code until merge (Helm-only git changes could still sync). Choose PR builds because the user wants to run a branch on Talos.

**templ/CSS:** keep generating on the laptop and committing (current `web.Dockerfile` `go build`s committed `_templ.go`). Optional later: generate inside the image so CI is hermetic.

### 5–8. Edge Gateway, no waypoint, Hubble L4, protocol ports, Cloudflare cutover

Unchanged in intent from the Cilium cutover (class `cilium`, in-cluster Gateway Service DNS — LoadBalancer type, not ClusterIP enum — drop Istio, Hubble L4, host-based HTTPRoutes).

## Risks / Trade-offs

| Risk                                                 | Mitigation                                                                                 |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Argo syncs git before PR images exist                | Pin `imageTag` to SHA; wait for the image workflow; or `ImageUpdater` later (not required) |
| Full image matrix on every PR push is slow/expensive | Path-filter matrix; skip unchanged images (document if deferred)                           |
| Delete GHCR while Argo still uses the SHA            | Cleanup only on PR closed, after root is back on `main`                                    |
| One cluster, two branches                            | Document: one live revision; do not ApplicationSet in this change                          |
| Tiltfile leftover confuses DX                        | Delete Tiltfile and `infra/argocd/local/` in this change                                   |
| Generated templ drift without Tilt watch             | devenv/hooks; commit generated files as today                                              |
| Unifying observability defaults grows RAM vs Colima  | Accepted: Talos is the only cluster                                                        |

## Migration Plan

1. Confirm Talos `GatewayClass/cilium`; Hubble up; Argo already on the cluster.
2. Add PR image workflow + SHA tags; add closed-PR GHCR cleanup.
3. Collapse to one Argo root; GHCR-required chart defaults; full observability defaults; CNPG in the marketplace release; delete Tiltfile, local-root, Colima devenv wiring.
4. Switch ingress to Cilium Gateway; update Cloudflare origin DNS; verify checkout.
5. Remove Istio apps/charts/scrapes and remaining Tilt/Istio labels.
6. OpenSpec validate; update #38.

Rollback: Istio Gateway + origin DNS as before. Images: keep `:main`; PR cleanup is independent.
