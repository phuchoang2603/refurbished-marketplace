# Cilium, Gateway API, and ingress

Cilium is cluster bootstrap from **talos-proxmox**, not an Argo Application in this repo. Empty-cluster CNI still needs that Helm install; this repo must not duplicate `apps/values/cilium.yaml` (a second copy would drift and can wipe mesh flags on `helm upgrade`).

Already on the cluster (verified against talos-dev Helm user values):

- CNI, kube-proxy replacement, L2 announcements, Gateway API, Hubble relay+UI
- WireGuard encryption, Envoy L7 proxy, `cluster.name=talos` / `cluster.id=1`
- L2 IP pool + platform Gateways in `cilium-ingress` (Hubble `10.69.100.1`, Longhorn, Argo CD)

This repo only consumes that dataplane: marketplace `Gateway`/`HTTPRoute` (`gatewayClassName: cilium`), Cloudflare origin DNS, and later `CiliumNetworkPolicy` if a follow-on change needs it. Do not add WireGuard/Envoy/ClusterMesh values here.

Marketplace browser traffic: Cloudflare Tunnel → Cilium Gateway API.

Cilium 1.18 Gateway Services are `LoadBalancer` (ClusterIP is not a valid `CiliumGatewayClassConfig` service type). `cloudflared` still uses in-cluster DNS:

`http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`

East–west traffic is ordinary ClusterIP. Hubble L4 is the network observe path; application traces stay OTEL → VictoriaTraces.

## GitOps

`talos-root` (`infra/argocd/talos/root.yaml`) renders [`infra/argocd/app-of-apps`](../../infra/argocd/app-of-apps/). Marketplace enrollment has no Istio labels. Kafka stays in namespace `kafka`.

## Edge

| Hostname             | Backend                     |
| -------------------- | --------------------------- |
| `shop.phuchoang.sbs` | `web`                       |
| `pay.phuchoang.sbs`  | `payment-gateway-simulator` |

HTTPRoutes set `X-Forwarded-Proto: https` and `X-Forwarded-Host` so hosted-payment callbacks are not rewritten to HTTP (POST → Cloudflare 301 → GET → 405).

Cloudflare Zero Trust Public Hostnames (not in Git):

- `shop.phuchoang.sbs` → `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`
- `pay.phuchoang.sbs` → `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`

TLS terminates at Cloudflare. No marketplace TLS Secret on the Gateway. Do not reuse the `cilium-ingress` Hubble/Longhorn/Argo Gateways for shop/pay.

```bash
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl get pods -n cloudflare-tunnel
```

## Hubble

On the LAN (cluster bootstrap, not this chart): `http://10.69.100.1` (dev). Optional:

```bash
kubectl -n kube-system port-forward svc/hubble-ui 12000:80
```

## Rollback

Disable marketplace `ingress.enabled` and sync, or repoint Cloudflare hostnames. Cilium itself stays with talos-proxmox.
