## Context

Today: Tilt on Colima owns marketplace Helm + `docker_build` + templ/tailwind watches; Argo `local-root` skips marketplace; staging Argo tracks `main` and GHCR `:main`. Talos already runs Cilium. Tilt against Talos would need a registry and `live_update` because Talos does not share the laptop Docker daemon.

New intent: **no Tilt deploy path.** Talos + Argo CD is the only applier. Dev and prod are two Talos clusters; git branch + GHCR SHA is the inner loop on talos-dev.

templ `_templ.go` files are already committed; Tailwind CSS is copied from the repo in `web.Dockerfile`. Watches were convenience, not a cluster requirement.

Follow-on: `add-cilium-mesh-policy-and-canary`.

## Goals / Non-Goals

**Goals:**

- One dataplane: Cilium CNI + Cilium Gateway API on Talos. Hubble is off; traces are app OTEL.
- Same browser contract: Cloudflare HTTPS, HTTP origin, host-based routes, forwarded proto/host headers.
- Dev and prod share one Talos chart shape (GHCR, observability, CNPG-in-chart, Cilium Gateway). Env is two clusters: Doppler token on the cluster, thin Argo roots (`dev-root` / `prod-root`), and `values-prod.yaml` for prod hostnames only.
- CI publishes images Argo can pull: SHA tags on `main` and on PRs; prune PR-only GHCR versions when the PR closes.
- Remove Istio; kafka stays out of L7 intercept.

**Non-Goals:**

- Tilt, Colima/k3s, Tiltfile workarounds, and a second “local” Helm personality.
- Preview environments per PR (extra namespaces, hosts, CNPG clusters). One live `ecommerce` on the cluster.
- Hubble L7 metrics / Istio-RED replacement (follow-on [#43](https://github.com/phuchoang2603/refurbished-marketplace/issues/43) app-level OTEL metrics).
- Cilium as an Argo Application.
- Mesh/Gateway proxy spans.

## Decisions

### 1. Cilium is cluster-owned in talos-proxmox; do not copy Helm values here

Source of truth: sibling repo **talos-proxmox** `apps/values/cilium.yaml` (installed by cluster bootstrap, chart 1.18.x). Marketplace GitOps only consumes `gatewayClassName: cilium`. Hubble may be disabled in those cluster values; this repo does not require it. L2 pool and platform Gateways (Longhorn, Argo CD) live in talos-proxmox network manifests.

This repo does **not** keep a second Cilium values file (it would drift). Marketplace GitOps only consumes `gatewayClassName: cilium`. Follow-on `add-cilium-mesh-policy-and-canary` may add policies/routes; it must not helm-upgrade Cilium.

### 2. Argo CD is the only deploy path; delete Tilt-era quirks

One app-of-apps catalog, two thin roots (`dev-root` / `prod-root`) on the **gpu** cluster. Children destine registered Argo clusters `dev` / `prod`. Chart **defaults** are talos-dev (`shop-dev` / `pay-dev`). Prod uses `values-prod.yaml` hosts. `dev-root` `targetRevision` stays on the branch we are running until we change it on purpose (not automatically back to `main` at merge).

**Delete (legacy / Tilt-only):**

| Quirk                                                                                                             | Why it existed                                   | Replacement                                                                |
| ----------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ | -------------------------------------------------------------------------- |
| `Tiltfile` (install Argo, Gateway API CRDs, Doppler secret, `local-root`, `docker_build`, watches, port-forwards) | Colima inner loop                                | Git push → CI → Argo; `kubectl` port-forward if needed                     |
| Strip Namespace from Helm + Tilt `ecommerce-namespace` labels                                                     | Tilt prune/recreate deleted CNPG                 | Argo `CreateNamespace` + `managedNamespaceMetadata` only (no Istio labels) |
| Out-of-band `helm template databases.tpl \| kubectl apply`                                                        | `tilt down` deleted CNPG PVCs / Debezium offsets | Marketplace chart owns CNPG Clusters; `tilt down` is gone                  |
| `local-root` with `marketplace.enabled: false`                                                                    | Tilt owned the app chart                         | Single root; marketplace on                                                |
| Empty `global.imageRegistry` → short image names                                                                  | Tilt loaded images into Colima Docker            | Always `ghcr.io/.../name:<sha>` or `:main`                                 |
| Marketplace `values.yaml` Colima CPU/mem / 512Mi PVCs                                                             | Fit 4 CPU / 8 GiB k3s                            | Defaults = current staging-class requests/limits/storage                   |
| Observability chart defaults apps-only (no node-exporter/ksm/Alertmanager/default dashboards)                     | Colima RAM                                       | Defaults = full platform stack (former `values-staging.yaml` shape)        |
| Dual Istio CNI `platform: k3s` vs RKE2 overlays                                                                   | Colima vs old staging                            | Istio charts deleted                                                       |
| `mesh.tpl` comments + ambient labels for Tilt vs Argo namespace fight                                             | Two appliers                                     | No waypoint; Argo owns ns metadata                                         |
| devenv `DOCKER_HOST` Colima k8s socket, `tilt` package                                                            | Local k8s + Tilt                                 | Docker only if Testcontainers needs it; drop Tilt                          |
| Tilt-applied `doppler-token.dev.secret.yaml`                                                                      | Local bootstrap                                  | Same pattern as remote: bootstrap Secret once, ESO `ClusterSecretStore`    |
| Tilt extra Gateway API CRD apply                                                                                  | Colima had no CRDs                               | Talos/Cilium already has Gateway API                                       |
| Docs/PR template/`CONTRIBUTING` `tilt up` on Colima                                                               | Old DX                                           | Argo + GHCR                                                                |
| `.dev` Cloudflare hosts as chart default                                                                          | Second Colima front door                         | Talos-dev uses `shop-dev`/`pay-dev`; prod overlay uses `shop`/`pay`        |

**Allowed to differ between “what’s live on main” and a branch:** `targetRevision`, `global.imageTag` (SHA via `$ARGOCD_APP_REVISION`), Doppler config if secrets must not mix. Not a second cluster shape.

**Rationale:** matching dev and prod means the same YAML path, not a miniature stack that only resembles prod after overlays.

**Alternatives considered:** keep Colima-sized defaults with a `values-talos.yaml` overlay (still two personalities); Tilt `live_update` (rejected).

### 3. Branch tracking is one pointer, not N preview envs

Default: `dev-root` `imageTag` = `$ARGOCD_APP_REVISION`, `prod-root` `imageTag: main`. Git `targetRevision` can still be a branch on talos-dev.

To run a feature on the cluster: point `dev-root` at the branch; wait for CI to push `:<sha>`. One `ecommerce` namespace — only one git revision live at a time.

Do **not** ApplicationSet-per-PR in this change (would need `pr-N` namespaces, extra shop hosts, extra CNPG).

**Rationale:** homelab has one shop. Branch tracking is “what is live,” not “every PR gets a URL.”

**Alternatives considered:** ApplicationSet PR generator (correct for multi-preview, too much here); only deploy after merge (simplest; then PR images are optional until you want to flip the pointer).

### 4. GHCR `:<sha>` plus `:main` on the default branch

- Same release matrix; always tag the git SHA (PR **head** SHA on pull_request).
- Tag `:main` only on `refs/heads/main`. No `:dev` or `pr-<n>`. Closed PRs delete SHA-only GHCR versions; never `:main`.
- Argo `imagePullPolicy: Always` so kubelets pick up a new `:main` digest; SHA tags are unique.

**templ/CSS:** keep generating on the laptop and committing (current `web.Dockerfile` `go build`s committed `_templ.go`). Optional later: generate inside the image so CI is hermetic.

### 5–8. Edge Gateway, no waypoint, traces without Hubble, protocol ports, Cloudflare cutover

Cilium Gateway class `cilium`. Origin Service is LoadBalancer (Cilium 1.18) but `cloudflared` uses in-cluster DNS `cilium-gateway-<gateway-name>.<ns>.svc.cluster.local:80`. No Istio waypoint. Service ports named `grpc` / `http` from chart `protocol`. Grafana has its own Gateway in `monitoring`. Hubble is not part of this observe path.

## Risks / Trade-offs

| Risk                                                 | Mitigation                                                        |
| ---------------------------------------------------- | ----------------------------------------------------------------- |
| Argo still pins a PR SHA when cleanup runs           | Retarget `dev-root` to `main` before or as the PR closes          |
| Full image matrix on every PR push is slow/expensive | Path-filter matrix; skip unchanged images (document if deferred)  |
| One cluster, two branches                            | Document: one live revision; do not ApplicationSet in this change |
| Tiltfile leftover confuses DX                        | Delete Tiltfile and `infra/argocd/local/` in this change          |
| Generated templ drift without Tilt watch             | devenv/hooks; commit generated files as today                     |
| Unifying observability defaults grows RAM vs Colima  | Accepted: Talos is the only cluster                               |

## Migration Plan

1. Confirm Talos `GatewayClass/cilium`; Argo on gpu destining `dev`/`prod`.
2. Add PR image workflow + SHA tags; add closed-PR GHCR cleanup.
3. Collapse to one Argo root; GHCR-required chart defaults; full observability defaults; CNPG in the marketplace release; delete Tiltfile, local-root, Colima devenv wiring.
4. Switch ingress to Cilium Gateway; update Cloudflare origin DNS; verify checkout.
5. Remove Istio apps/charts/scrapes and remaining Tilt/Istio labels.
6. OpenSpec validate; update #38.

Rollback: Istio Gateway + origin DNS as before. Images: keep `:main`; PR cleanup is independent.
