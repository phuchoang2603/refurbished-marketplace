# Contributing

Thanks for helping build this project. This guide is the entry point; detailed how-tos live under [docs/development/](docs/development/) and [docs/deployment/](docs/deployment/).

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A local Kubernetes runtime (for example Colima + Docker / k3s)
- [Doppler](https://www.doppler.com/) account for cluster secrets
- Cloudflare Zero Trust tunnel token in Doppler `dev` for local `.dev` hostnames

## Quick start

```bash
devenv shell
# one-time secrets + Cloudflare Public Hostnames — see docs/development/local-setup.md
bootstrap-local-argocd
build-images
# browse https://shop.dev.phuchoang.sbs
```

## Development guides

Local workflow, codegen, and project conventions.

| Topic                          | Guide                                                                      |
| ------------------------------ | -------------------------------------------------------------------------- |
| devenv shell, local Argo, test | [docs/development/local-setup.md](docs/development/local-setup.md)         |
| Doppler + External Secrets     | [docs/development/secrets.md](docs/development/secrets.md)                 |
| Code generation and formatting | [docs/development/code-generation.md](docs/development/code-generation.md) |
| OpenSpec planning workflow     | [docs/development/openspec.md](docs/development/openspec.md)               |
| Issues, labels, pull requests  | [docs/development/github-workflow.md](docs/development/github-workflow.md) |

## Deployment guides

CI, container releases, and remote cluster GitOps.

| Topic                            | Guide                                                  |
| -------------------------------- | ------------------------------------------------------ |
| GitHub Actions, GHCR releases    | [docs/deployment/ci.md](docs/deployment/ci.md)         |
| Argo CD GitOps (local + staging) | [docs/deployment/gitops.md](docs/deployment/gitops.md) |

## Questions

Open a **Feature**, **Bug**, or **Chore** issue. For multi-step work, use one Feature issue with a checklist or split into several linked issues — see [docs/development/github-workflow.md](docs/development/github-workflow.md).
