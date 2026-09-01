# Local setup

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- Talos kubeconfigs: `~/.kube/talos-gpu.yaml` (Argo CD), `~/.kube/talos-dev.yaml` (workloads / Doppler)
- [Doppler](https://doppler.com/) — see [secrets.md](secrets.md)
- Cloudflare Zero Trust tunnel for `shop-dev.phuchoang.sbs` / `pay-dev.phuchoang.sbs`

Argo CD runs on the **gpu** cluster (talos-proxmox). Workloads sync to the registered Argo cluster named `dev`. Bootstrap Doppler **dev** on talos-dev, then apply the **dev** root on gpu:

```bash
export KUBECONFIG="$HOME/.kube/talos-dev.yaml"
kubectl create namespace operators --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.dev.secret.yaml

export KUBECONFIG="$HOME/.kube/talos-gpu.yaml"
kubectl apply -f infra/argocd/dev/root.yaml
```

`prod-root` is the same pattern with the prod kubeconfig for Doppler and `infra/argocd/prod/root.yaml` on gpu. Do not apply the `prd` Doppler token on talos-dev.

Children follow the root’s git revision (`spec.source.targetRevision` in `infra/argocd/dev/root.yaml`). Change that field in git, then `kubectl apply -f` the same file **on gpu**. `global.imageTag` is `$ARGOCD_APP_REVISION`. Wait for GHCR `:<sha>` after the image job. Prod uses `:main`. Before this PR merges, set `targetRevision` back to `main`.

## Development shell

```bash
devenv shell
```

Go, protobuf, `kubectl`, `helm`, Doppler, OpenSpec. On enter, devenv regenerates proto/sqlc/templ/Tailwind when those inputs change (`codegen:templ`, `codegen:tailwind`). Commit the generated files; CI builds the `web` image from them.

## Browser

Cloudflare Tunnel → Cilium Gateway (`cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`).

| Hostname                 | Backend                     |
| ------------------------ | --------------------------- |
| `shop-dev.phuchoang.sbs` | `web`                       |
| `pay-dev.phuchoang.sbs`  | `payment-gateway-simulator` |

Smoke-check:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-gpu.yaml" get applications -n argo-cd
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get gateway,httproute -n ecommerce
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get pods -n ecommerce
```

Optional debug: `kubectl -n ecommerce port-forward svc/web 8080:8080`.

## Integration testing

Integration tests use Testcontainers (Docker on the laptop). Full flows: Talos + Argo + GHCR.
