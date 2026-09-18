# Local setup

Runtime layout is in [architecture.md](../architecture.md). This page is how to boot the Talos/Argo loop on a laptop.

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- Talos kubeconfigs: `~/.kube/talos-argocd.yaml` (management cluster / Argo CD), `~/.kube/talos-dev.yaml` (workloads / Doppler)
- [Doppler](https://doppler.com/) — see [secrets.md](secrets.md)
- Cloudflare Zero Trust tunnel for `shop-dev.phuchoang.sbs` / `pay-dev.phuchoang.sbs`

Argo CD runs on the **management** cluster (talos-proxmox). Workloads sync to the registered Argo cluster named `dev`. Apply the `platform-dev` root first, bootstrap Doppler **dev** on talos-dev, verify the platform prerequisites, then apply the marketplace **dev** root:

```bash
export KUBECONFIG="$HOME/.kube/talos-dev.yaml"
kubectl create namespace operators --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.dev.secret.yaml

export KUBECONFIG="$HOME/.kube/talos-argocd.yaml"
kubectl apply --server-side -f ../talos-proxmox/apps/argocd/roots/dev.yaml

export KUBECONFIG="$HOME/.kube/talos-dev.yaml"
kubectl wait --for=condition=Ready secretstore/doppler -n ecommerce --timeout=2m

export KUBECONFIG="$HOME/.kube/talos-argocd.yaml"
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply -f infra/argocd/dev/root.yaml
```

`prod-root` is the same pattern with the prod kubeconfig for Doppler and `infra/argocd/prod/root.yaml` on the management cluster. Do not apply the `prd` Doppler token on talos-dev.

Children follow the root’s git revision (`spec.source.targetRevision` in `infra/argocd/dev/root.yaml`), which is committed as `main`. `global.imageTag` is `$ARGOCD_APP_REVISION`; wait for every required GHCR `:<sha>` image before syncing dev. Prod uses `:main`.

## Development shell

```bash
devenv shell
```

Go, protobuf, `kubectl`, `helm`, Doppler, OpenSpec. On enter, devenv runs tasks to regenerate proto/sqlc/templ/Tailwind and sync modules when inputs change (`codegen:proto`, `codegen:sqlc`, `codegen:templ`, `codegen:tailwind`, `go:tidy`). Run any task directly with `devenv tasks run <task>`. Commit the generated files; CI builds the `web` image from them.

## Browser

Cloudflare Tunnel → Cilium Gateway (`cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`).

| Hostname                 | Backend                     |
| ------------------------ | --------------------------- |
| `shop-dev.phuchoang.sbs` | `web`                       |
| `pay-dev.phuchoang.sbs`  | `payment-gateway-simulator` |

Smoke-check:

```bash
kubectl --kubeconfig="$HOME/.kube/talos-argocd.yaml" get applications -n argo-cd
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get gateway,httproute -n ecommerce
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get svc -n ecommerce -l gateway.networking.k8s.io/gateway-name=ecommerce-ingress
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" get pods -n ecommerce
```

Optional debug: `kubectl -n ecommerce port-forward svc/web 8080:8080`.

## Integration testing

Integration tests use Testcontainers (Docker on the laptop). Full flows: Talos + Argo + GHCR.

## Infrastructure formatting

The existing treefmt/Oxfmt hook formats plain YAML, including chart values and Argo roots. Helm template files remain excluded. No Helm validation script or devenv validation task is configured.
