## 1. Cilium values and Argo-only DX

- [ ] 1.1 Add documented Cilium Helm values for Talos (cluster-owned; match the running install).
- [ ] 1.2 One Talos Argo root with marketplace enabled; delete Tilt as applier and `marketplace.enabled: false` local-root.
- [ ] 1.3 Spike Cilium Gateway ClusterIP (`CiliumGatewayClassConfig` or equivalent) and `HTTPRoute` `RequestHeaderModifier`; record the origin Service DNS name.

## 2. GHCR PR images and cleanup

- [ ] 2.1 Add a pull_request image build that pushes `:<git-sha>` (and optional `pr-<n>`) without moving `:main`.
- [ ] 2.2 Add pull_request closed cleanup that deletes PR-only GHCR versions and never deletes `:main` or a SHA Argo still pins.
- [ ] 2.3 Document: wait for image job before Argo sync; retarget root to `main` before cleanup; path-filter the matrix if full builds are too heavy.

## 3. Marketplace ingress cutover

- [ ] 3.1 Switch `ingress.tpl` to a Cilium GatewayClass; drop Istio ClusterIP annotation; keep host-based HTTPRoutes and X-Forwarded-* filters.
- [ ] 3.2 Render GatewayClass parameters / ClusterIP config from Git so `cloudflared` can use in-cluster Service DNS without an L2 VIP.
- [ ] 3.3 Delete `mesh.tpl`, `mesh.*` values, and Argo `istio.io/*` namespace labels.
- [ ] 3.4 Update Cloudflare tunnel comments/docs to the new Gateway Service DNS (not `ecommerce-ingress-istio`).

## 4. Remove Istio from GitOps

- [ ] 4.1 Remove Istio apps from `infra/argocd/app-of-apps` and staging-root `valueFiles` for istiod/cni/ztunnel.
- [ ] 4.2 Delete `infra/charts/operators/istio/` wrapper charts.
- [ ] 4.3 Drop istio-system privileged PSS once Istio apps are gone; keep monitoring PSS for node-exporter.

## 5. Observability

- [ ] 5.1 Remove `istioScrapes` templates/values and the Marketplace Istio RED dashboard.
- [ ] 5.2 Drop `istio-proxy` log exclude filters if those containers no longer exist.
- [ ] 5.3 Document Hubble L4 (UI/port-forward) as the network observe path; leave app OTEL → VictoriaTraces unchanged.

## 6. Docs and Tilt-legacy removal

- [ ] 6.1 Replace `docs/deployment/istio.md` with Cilium + Gateway API + Cloudflare origin docs.
- [ ] 6.2 Rewrite gitops/local-setup/secrets/code-generation/CONTRIBUTING/PR template/README: Argo + GHCR only; no Colima k8s, no `tilt up`.
- [ ] 6.3 Delete `Tiltfile` and `infra/argocd/local/`.
- [ ] 6.4 Marketplace chart: drop empty `imageRegistry` / Colima resource+PVC defaults; CNPG stays in the Helm release; shop/pay hosts as chart default (no `.dev` profile).
- [ ] 6.5 Observability chart: make full platform stack the default; delete apps-only Colima defaults / dual PVC sizes.
- [ ] 6.6 devenv: remove `tilt` package and Colima `DOCKER_HOST` k8s socket wiring (keep Docker only if Testcontainers needs it).
- [ ] 6.7 Stop Tilt-applying Doppler token and Gateway API CRDs.

## 7. Verify and close

- [ ] 7.1 Argo on Talos: Cilium Gateway Accepted, Cloudflare hosts, checkout (callback not 405).
- [ ] 7.2 PR: images appear on GHCR at the head SHA; optional branch track on the cluster; after close, PR-only tags gone and `:main` intact.
- [ ] 7.3 Run `openspec validate replace-istio-with-cilium` and update GitHub issue [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38).
