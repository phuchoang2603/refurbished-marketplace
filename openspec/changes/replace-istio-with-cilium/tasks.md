## 1. Cilium values and Argo-only DX

- [x] 1.1 Point at talos-proxmox Cilium Helm values (cluster-owned; do not duplicate in this repo).
- [x] 1.2 One Talos Argo root with marketplace enabled; delete Tilt as applier and `marketplace.enabled: false` local-root.
- [x] 1.3 Spike Cilium Gateway ClusterIP (`CiliumGatewayClassConfig` or equivalent) and `HTTPRoute` `RequestHeaderModifier`; record the origin Service DNS name.

## 2. GHCR SHA + main

- [x] 2.1 Tag `:<git-sha>` on every image build; add `:main` only on `refs/heads/main`.
- [x] 2.2 Dev-root `imageTag` = `$ARGOCD_APP_REVISION`; prod-root `imageTag: main`; no PR-tag cleanup workflow.
- [x] 2.3 Document wait for GHCR `:<sha>` before expecting pods to pull.

## 3. Marketplace ingress cutover

- [x] 3.1 Switch `ingress.tpl` to a Cilium GatewayClass; drop Istio ClusterIP annotation; keep host-based HTTPRoutes and X-Forwarded-* filters.
- [x] 3.2 Render GatewayClass parameters / ClusterIP config from Git so `cloudflared` can use in-cluster Service DNS without an L2 VIP.
- [x] 3.3 Delete `mesh.tpl`, `mesh.*` values, and Argo `istio.io/*` namespace labels.
- [x] 3.4 Update Cloudflare tunnel comments/docs to the new Gateway Service DNS (not `ecommerce-ingress-istio`).

## 4. Remove Istio from GitOps

- [x] 4.1 Remove Istio apps from `infra/argocd/app-of-apps` and staging-root `valueFiles` for istiod/cni/ztunnel.
- [x] 4.2 Delete `infra/charts/operators/istio/` wrapper charts.
- [x] 4.3 Drop istio-system privileged PSS once Istio apps are gone; keep monitoring PSS for node-exporter.

## 5. Observability

- [x] 5.1 Remove `istioScrapes` templates/values and the Marketplace Istio RED dashboard.
- [x] 5.2 Drop `istio-proxy` log exclude filters if those containers no longer exist.
- [x] 5.3 Document Hubble L4 (UI/port-forward) as the network observe path; leave app OTEL → VictoriaTraces unchanged.

## 6. Docs and Tilt-legacy removal

- [x] 6.1 Replace `docs/deployment/istio.md` with Cilium + Gateway API + Cloudflare origin docs.
- [x] 6.2 Rewrite gitops/local-setup/secrets/code-generation/CONTRIBUTING/PR template/README: Argo + GHCR only; no Colima k8s, no `tilt up`.
- [x] 6.3 Delete `Tiltfile` and `infra/argocd/local/`.
- [x] 6.4 Marketplace chart: drop empty `imageRegistry` / Colima resource+PVC defaults; CNPG stays in the Helm release; shop/pay hosts as chart default (no `.dev` profile).
- [x] 6.5 Observability chart: make full platform stack the default; delete apps-only Colima defaults / dual PVC sizes.
- [x] 6.6 devenv: remove `tilt` package and Colima `DOCKER_HOST` k8s socket wiring (keep Docker only if Testcontainers needs it).
- [x] 6.7 Stop Tilt-applying Doppler token and Gateway API CRDs.

## 7. Verify and close

- [ ] 7.1 Argo on Talos: Cilium Gateway Accepted, Cloudflare hosts, checkout (callback not 405).
- [ ] 7.2 PR: GHCR `:<head-sha>` exists; talos-dev `$ARGOCD_APP_REVISION` matches; `:main` unchanged until merge to main.
- [x] 7.3 Run `openspec validate replace-istio-with-cilium` and update GitHub issue [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38).
