## 1. Cilium values and Colima bootstrap

- [ ] 1.1 Add documented Cilium Helm values for Talos (cluster-owned; match the running install: kube-proxy replacement, Gateway API, Hubble, Talos cgroup/devices).
- [ ] 1.2 Add Colima/k3s Cilium values (non-Talos `k8sServiceHost`/`Port`, cgroup, devices) and local-setup steps so k3s runs Cilium instead of default CNI and `GatewayClass/cilium` is Accepted.
- [ ] 1.3 Spike Cilium Gateway ClusterIP (`CiliumGatewayClassConfig` or equivalent) and `HTTPRoute` `RequestHeaderModifier`; record the origin Service DNS name.

## 2. Marketplace ingress cutover

- [ ] 2.1 Switch `ingress.tpl` to a Cilium GatewayClass; drop Istio ClusterIP annotation; keep host-based HTTPRoutes and X-Forwarded-* filters.
- [ ] 2.2 Render GatewayClass parameters / ClusterIP config from Git so `cloudflared` can use in-cluster Service DNS without an L2 VIP.
- [ ] 2.3 Delete `mesh.tpl`, `mesh.*` values, Argo `istio.io/*` namespace labels, and Tilt ambient/waypoint label apply.
- [ ] 2.4 Update Cloudflare tunnel comments/docs to the new Gateway Service DNS (not `ecommerce-ingress-istio`).

## 3. Remove Istio from GitOps

- [ ] 3.1 Remove Istio apps from `infra/argocd/app-of-apps` and staging-root `valueFiles` for istiod/cni/ztunnel.
- [ ] 3.2 Delete `infra/charts/operators/istio/` wrapper charts.
- [ ] 3.3 Drop istio-system privileged PSS once Istio apps are gone; keep monitoring PSS for node-exporter.

## 4. Observability

- [ ] 4.1 Remove `istioScrapes` templates/values and the Marketplace Istio RED dashboard.
- [ ] 4.2 Drop `istio-proxy` log exclude filters if those containers no longer exist.
- [ ] 4.3 Document Hubble L4 (UI/port-forward) as the network observe path; leave app OTEL → VictoriaTraces unchanged.

## 5. Docs

- [ ] 5.1 Replace `docs/deployment/istio.md` with Cilium + Gateway API + Cloudflare origin docs (TLS ownership, rollback, kafka namespace).
- [ ] 5.2 Update `docs/deployment/gitops.md`, `docs/development/local-setup.md`, and README Istio references.

## 6. Verify and close

- [ ] 6.1 Local: Cilium Gateway Accepted, Cloudflare `.dev` hosts, product → cart → checkout → payment (callback not 405).
- [ ] 6.2 Staging: same flow on shop/pay hosts; Hubble L4 flows visible; Grafana checkout TraceId still connected without waypoint/ingress spans.
- [ ] 6.3 Run `openspec validate replace-istio-with-cilium` and update GitHub issue [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38).
