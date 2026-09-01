# Contributing

Thanks for helping build this project. Guides live under [docs/development/](docs/development/) and [docs/deployment/](docs/deployment/).

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/)
- Talos kubeconfig
- [Doppler](https://www.doppler.com/) for cluster secrets
- Cloudflare tunnel for shop/pay hostnames

## Quick start

```bash
devenv shell
export KUBECONFIG="$HOME/.kube/talos-dev.yaml"
# secrets: docs/development/secrets.md
kubectl apply -f infra/argocd/talos/root.yaml
# https://shop.phuchoang.sbs
```

## Development guides

| Topic                      | Guide                                                                      |
| -------------------------- | -------------------------------------------------------------------------- |
| devenv, Argo, test         | [docs/development/local-setup.md](docs/development/local-setup.md)         |
| Doppler + External Secrets | [docs/development/secrets.md](docs/development/secrets.md)                 |
| Code generation            | [docs/development/code-generation.md](docs/development/code-generation.md) |
| OpenSpec                   | [docs/development/openspec.md](docs/development/openspec.md)               |
| GitHub issues / PRs        | [docs/development/github-workflow.md](docs/development/github-workflow.md) |

## Deployment guides

| Topic                         | Guide                                                  |
| ----------------------------- | ------------------------------------------------------ |
| GitHub Actions, GHCR          | [docs/deployment/ci.md](docs/deployment/ci.md)         |
| Argo CD GitOps                | [docs/deployment/gitops.md](docs/deployment/gitops.md) |
| Cilium + Gateway + Cloudflare | [docs/deployment/cilium.md](docs/deployment/cilium.md) |
