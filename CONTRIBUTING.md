# Contributing

Thanks for helping build this project. Guides live under [docs/](docs/), including [architecture](docs/architecture.md), [docs/development/](docs/development/), and [docs/deployment/](docs/deployment/).

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/)
- The Talos prod cluster brought up by `talos-proxmox` (it runs Argo CD)
- [Doppler](https://www.doppler.com/) access to `talos-proxmox` (kubeconfig) and `refurbished-marketplace` (app secrets)
- Cloudflare tunnel for the shop hostname

## Quick start

```bash
devenv shell
doppler secrets get KUBECONFIG --plain --project talos-proxmox --config prod \
  > "$HOME/.kube/talos-prod.yaml"
export KUBECONFIG="$HOME/.kube/talos-prod.yaml"
kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
kubectl apply --server-side -f infra/argocd/project.yaml
kubectl apply --server-side -f infra/argocd/prod/root.yaml
# https://shop.phuchoang.sbs
```

## Development guides

| Topic                      | Guide                                                                      |
| -------------------------- | -------------------------------------------------------------------------- |
| devenv, Argo, test         | [docs/development/local-setup.md](docs/development/local-setup.md)         |
| Architecture               | [docs/architecture.md](docs/architecture.md)                               |
| Doppler + External Secrets | [docs/development/secrets.md](docs/development/secrets.md)                 |
| Code generation            | [docs/development/code-generation.md](docs/development/code-generation.md) |
| OpenSpec                   | [docs/development/openspec.md](docs/development/openspec.md)               |
| GitHub issues / PRs        | [docs/development/github-workflow.md](docs/development/github-workflow.md) |

## Deployment guides

| Topic                                 | Guide                                                                |
| ------------------------------------- | -------------------------------------------------------------------- |
| GitHub Actions, GHCR                  | [docs/deployment/ci.md](docs/deployment/ci.md)                       |
| Argo CD GitOps                        | [docs/deployment/gitops.md](docs/deployment/gitops.md)               |
| Istio ambient + Gateway + Cloudflare  | [docs/deployment/networking.md](docs/deployment/networking.md)       |
| Observability (Grafana, traces, logs) | [docs/deployment/observability.md](docs/deployment/observability.md) |
