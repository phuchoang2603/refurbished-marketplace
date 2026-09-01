# Local setup

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- Talos kubeconfig (for example `KUBECONFIG=$HOME/.kube/talos-dev.yaml`)
- [Doppler](https://doppler.com/) — see [secrets.md](secrets.md)
- Cloudflare Zero Trust tunnel for `shop.phuchoang.sbs` / `pay.phuchoang.sbs`

Argo CD must already be installed on the cluster. Apply the GitOps root from this repo:

```bash
export KUBECONFIG="$HOME/.kube/talos-dev.yaml"
kubectl apply -f infra/argocd/talos/root.yaml
```

Children follow the root’s git revision. Push the branch, wait for GHCR image jobs (`:<sha>`), then set `global.imageTag` on the root to that SHA if you are not on `main`.

## Development shell

```bash
devenv shell
```

Go, protobuf, `kubectl`, `helm`, Doppler, OpenSpec. On enter, devenv regenerates proto/sqlc when those inputs change.

templ / Tailwind: generate on the laptop and commit (`templ generate`, `tailwindcss …`). Images are built in GitHub Actions, not locally into the cluster.

## Browser

Cloudflare Tunnel → Cilium Gateway (`cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`).

| Hostname             | Backend                     |
| -------------------- | --------------------------- |
| `shop.phuchoang.sbs` | `web`                       |
| `pay.phuchoang.sbs`  | `payment-gateway-simulator` |

Smoke-check:

```bash
kubectl get applications -n argo-cd
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl get pods -n ecommerce
```

Optional debug: `kubectl -n ecommerce port-forward svc/web 8080:8080`.

## Integration testing

Integration tests use Testcontainers (Docker on the laptop). Full flows: Talos + Argo + GHCR.
