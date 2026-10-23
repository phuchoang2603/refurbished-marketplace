# Local setup

Runtime layout is in [architecture.md](../architecture.md). This page is how to boot the Talos/Argo loop on a laptop.

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A talos-prod cluster brought up by `talos-proxmox` (its platform root installs Argo CD)
- [Doppler](https://doppler.com/) access to projects `talos-proxmox` (kubeconfig) and `refurbished-marketplace` (app secrets) — see [secrets.md](secrets.md)
- Cloudflare Zero Trust tunnel for `shop.phuchoang.sbs`

Argo CD runs inside talos-prod and deploys only to that cluster. Fetch its kubeconfig from Doppler, apply the Doppler `prd` token, then the marketplace project and `prod-root`. Children retry until the platform operators and CRDs are available, so no readiness wait is needed:

```bash
umask 077
doppler secrets get KUBECONFIG --plain --project talos-proxmox --config prod \
  > "$HOME/.kube/talos-prod.yaml"
export KUBECONFIG="$HOME/.kube/talos-prod.yaml"

kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply --server-side -f infra/argocd/prod/root.yaml
```

The `prod` cluster is the development server. Children follow the root’s git revision (`spec.source.targetRevision` in `infra/argocd/prod/root.yaml`) and use the image SHA pinned in each chart’s `values-prod.yaml`. See [gitops.md](../deployment/gitops.md) for deployment configuration.

## Development shell

```bash
devenv shell
```

Go, protobuf, `kubectl`, `helm`, Doppler, OpenSpec. On enter, devenv runs tasks to regenerate proto/sqlc/templ/Tailwind and sync modules when inputs change (`codegen:proto`, `codegen:sqlc`, `codegen:templ`, `codegen:tailwind`, `go:tidy`). Run any task directly with `devenv tasks run <task>`. Commit the generated files; CI builds the `web` image from them.

## Browser

Cloudflare Tunnel → Istio Gateway (`ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`).

| Hostname             | Backend |
| -------------------- | ------- |
| `shop.phuchoang.sbs` | `web`   |

The prod overlay configures the external payment gateway. Chart defaults leave it disabled; checkout submission then returns unavailable without creating an order or reserving stock.

Smoke-check:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get applications -n argo-cd
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get gateway,httproute -n ecommerce
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" get pods -n ecommerce
```

Optional debug: `kubectl -n ecommerce port-forward svc/web 8080:8080`.

## Integration testing

Integration tests use Testcontainers (Docker on the laptop). Full flows: Talos + Argo + GHCR.

## Infrastructure formatting

The existing treefmt/Oxfmt hook formats plain YAML, including chart values and Argo roots. Helm template files remain excluded. No Helm validation script or devenv validation task is configured.
