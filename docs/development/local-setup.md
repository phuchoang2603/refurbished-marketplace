# Local setup

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- Talos kubeconfig (for example `KUBECONFIG=$HOME/.kube/talos-dev.yaml`)
- [Doppler](https://doppler.com/) — see [secrets.md](secrets.md)
- Cloudflare Zero Trust tunnel for `shop-dev.phuchoang.sbs` / `pay-dev.phuchoang.sbs`

Argo CD must already be installed on the cluster. On **talos-dev**, bootstrap Doppler **dev** then apply the **dev** root:

```bash
export KUBECONFIG="$HOME/.kube/talos-dev.yaml"
kubectl create namespace operators --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.dev.secret.yaml
kubectl apply -f infra/argocd/dev/root.yaml
```

Do not apply `prod-root` or the `prd` Doppler token on this cluster. Prod is a separate Talos kubecontext + `infra/argocd/prod/root.yaml`.

Children follow the root’s git revision (`spec.source.targetRevision` in `infra/argocd/dev/root.yaml`). Change that field in git, then `kubectl apply -f` the same file. `global.imageTag` is `$ARGOCD_APP_REVISION`. Wait for GHCR `:<sha>` after the image job. Prod uses `:main`. Before this PR merges, set `targetRevision` back to `main`.

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
kubectl get applications -n argo-cd
kubectl get gateway,httproute -n ecommerce
kubectl get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl get pods -n ecommerce
```

Optional debug: `kubectl -n ecommerce port-forward svc/web 8080:8080`.

## Integration testing

Integration tests use Testcontainers (Docker on the laptop). Full flows: Talos + Argo + GHCR.
