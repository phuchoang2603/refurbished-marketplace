# Cilium, Gateway API, and ingress

Cilium is cluster bootstrap from **talos-proxmox**, not an Argo Application in this repo. Empty-cluster CNI still needs that Helm install; this repo must not duplicate `apps/values/cilium.yaml` (a second copy would drift and can wipe mesh flags on `helm upgrade`).

Already on the cluster (Cilium Helm from talos-proxmox):

- CNI, kube-proxy replacement, L2 announcements, Gateway API
- WireGuard encryption, Envoy L7 proxy, cluster name/id as set in that repo
- L2 IP pool + platform Gateways in `cilium-ingress` (Longhorn, Argo CD; Hubble UI is not required and may be absent)

This repo only consumes that dataplane: marketplace and Grafana `Gateway`/`HTTPRoute` (`gatewayClassName: cilium`) and Cloudflare origin DNS. Do not add WireGuard/Envoy/ClusterMesh values here.

Marketplace browser traffic: Cloudflare Tunnel → Cilium Gateway API.

Cilium 1.18 Gateway Services are `LoadBalancer`. `cloudflared` uses in-cluster DNS (not the L2 VIP):

`http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`

East–west traffic is ordinary ClusterIP. Application traces stay OTEL → VictoriaTraces. Hubble is not part of the observe path; request/error/duration SLIs are follow-on issue [#43](https://github.com/phuchoang2603/refurbished-marketplace/issues/43).

## GitOps

`dev-root` / `prod-root` (`infra/argocd/dev/root.yaml`, `infra/argocd/prod/root.yaml`) render [`infra/argocd/app-of-apps`](../../infra/argocd/app-of-apps/). Marketplace enrollment has no Istio labels. Kafka stays in namespace `kafka`.

## Edge

| Env  | Hostname                    | Backend                     |
| ---- | --------------------------- | --------------------------- |
| dev  | `shop-dev.phuchoang.sbs`    | `web`                       |
| dev  | `pay-dev.phuchoang.sbs`     | `payment-gateway-simulator` |
| dev  | `grafana-dev.phuchoang.sbs` | `observability-grafana`     |
| prod | `shop.phuchoang.sbs`        | `web`                       |
| prod | `pay.phuchoang.sbs`         | `payment-gateway-simulator` |
| prod | `grafana.phuchoang.sbs`     | `observability-grafana`     |

HTTPRoutes set `X-Forwarded-Proto: https` and `X-Forwarded-Host` so hosted-payment callbacks are not rewritten to HTTP (POST → Cloudflare 301 → GET → 405).

Cloudflare Zero Trust Public Hostnames (not in Git):

- `shop-dev.phuchoang.sbs` / `pay-dev.phuchoang.sbs` (dev) → `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`
- `grafana-dev.phuchoang.sbs` (dev) → `http://cilium-gateway-grafana.monitoring.svc.cluster.local:80`
- `shop.phuchoang.sbs` / `pay.phuchoang.sbs` (prod) → same origin DNS on the prod cluster
- `grafana.phuchoang.sbs` (prod) → `http://cilium-gateway-grafana.monitoring.svc.cluster.local:80`

TLS terminates at Cloudflare. No marketplace TLS Secret on the Gateway. Do not reuse the `cilium-ingress` Longhorn/Argo Gateways for shop/pay.

```bash
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl get pods -n cloudflare-tunnel
```

## Rollback

Disable marketplace `ingress.enabled` and sync, or repoint Cloudflare hostnames. Cilium itself stays with talos-proxmox.
