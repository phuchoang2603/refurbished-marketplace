## Why

Issue [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38): Istio ambient is four GitOps charts, a waypoint, Istio CNI, and L7 scrapes for an observe-only edge. The Talos cluster already runs Cilium. Replacing Istio with Cilium, dropping Tilt as a deploy path, and delivering marketplace workloads only through Argo CD (git revision + GHCR images) leaves one dataplane and one delivery loop.

## What Changes

- **BREAKING:** Remove Istio wrapper charts and Argo Applications (`base`, `istiod`, `cni`, `ztunnel`). Marketplace traffic no longer uses `gatewayClassName: istio` or `istio-waypoint`.
- Switch marketplace ingress to Kubernetes Gateway API with `gatewayClassName: cilium`. Keep the same host-based HTTPRoutes, `X-Forwarded-Proto` / `X-Forwarded-Host` filters, and Cloudflare Tunnel as the HTTPS front door.
- Provision the Cilium Gateway origin as ClusterIP (or equivalent in-cluster DNS) so `cloudflared` does not depend on L2 announcement VIPs. Repoint Cloudflare Public Hostnames to the new Service DNS.
- Delete ambient namespace labels, waypoint Gateway (`mesh.tpl`), Tilt Istio labels, and Istio L7 VMPodScrapes / Marketplace Istio RED dashboard.
- Treat Hubble L4 plus existing app OTEL → VictoriaTraces as the observe path. No Cilium L7 visibility policies, no Hubble HTTP metrics requirement, no replacement Grafana RED dashboard in this change.
- Document expected Cilium Helm values for Talos (cluster-owned CNI, not Argo).
- Dev on Talos SHALL use the same Argo app-of-apps + GHCR path as prod, with overlays limited to real env deltas (git revision, image SHA vs `:main`, secrets/hostnames if they must differ).
- CI tags GHCR `:<git-sha>` on every image build and `:main` only on `refs/heads/main`. After the PR closes, delete those SHA package versions; keep `:main`.

- templ/Tailwind: generate in devenv and commit (or in-Dockerfile); no Tilt watches.
- Keep kafka in its own namespace with no L7 Cilium policies. Keep protocol-aware Service port names / `appProtocol`.
- Rewrite Istio docs into Cilium/Gateway docs; rewrite local-setup around Argo + GHCR, not Colima/Tilt.

## Capabilities

### New Capabilities

- `cilium-ingress`: GitOps-managed Cilium Gateway API edge for marketplace browser traffic (and simulator), Cloudflare Tunnel HTTP origin, ClusterIP origin, TLS ownership, and ingress rollback.
- `cilium-observability`: Hubble L4 plus app OTEL as the post-Istio observe path; no waypoint, no Istio scrapes, no L7 mesh metrics requirement.

### Modified Capabilities

- `istio-ingress`: Retired. All Istio Gateway requirements are removed in favor of `cilium-ingress`.
- `istio-observability`: Retired. Ambient, waypoint, Istio chart pins, and Istio L7 telemetry requirements are removed. Protocol-aware Service ports move to `cilium-observability`.
- `argocd-gitops`: Drop GitOps-managed Istio and the Tilt/Colima split (`local-root` without marketplace, empty GHCR, dual apply). One Talos root; marketplace always Argo; branch/`main` tracking; `imageTag` = git SHA on dev, `main` on prod.
- `platform-observability`: Stop Istio scrapes/RED. Chart defaults SHALL be the full platform stack (not Colima apps-only). One PVC/resource profile, not local-vs-staging sizes.
- `distributed-tracing`: Keep “no mesh/Gateway proxy spans in the waterfall”; drop Istio-specific wording.
- `ghcr-release`: `:<git-sha>` plus `:main` on the default branch. No Tilt `docker_build` / short image names.
- `external-secrets`: Bootstrap Doppler token like the remote cluster (not Tilt `kubectl apply` of `doppler-token.dev.secret.yaml`).

## Impact

- Deletes Tiltfile, `infra/argocd/local/`, Colima-oriented chart defaults, and devenv Tilt/Colima Docker wiring.
- Does not change Go service business logic, protobuf, databases, Kafka topics, or OTEL bootstrap.
- Follow-on mesh hardening and Argo canaries live in `add-cilium-mesh-policy-and-canary` ([#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39)).
