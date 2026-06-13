# Contributing

Thanks for helping build this project. This guide is the entry point; detailed how-tos live under [docs/development/](docs/development/).

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A local Kubernetes runtime used by Tilt (for example Colima + Docker)
- [Doppler](https://www.doppler.com/) account for cluster secrets

## Quick start

```bash
devenv shell
# one-time secrets setup — see docs/development/secrets.md
tilt up
```

## Development guides

| Topic                               | Guide                                                                      |
| ----------------------------------- | -------------------------------------------------------------------------- |
| devenv shell, Tilt, testing         | [docs/development/local-setup.md](docs/development/local-setup.md)         |
| Doppler + External Secrets Operator | [docs/development/secrets.md](docs/development/secrets.md)                 |
| Code generation and formatting      | [docs/development/code-generation.md](docs/development/code-generation.md) |
| OpenSpec planning workflow          | [docs/development/openspec.md](docs/development/openspec.md)               |
| Issues, labels, pull requests       | [docs/development/github-workflow.md](docs/development/github-workflow.md) |
| CI, path filters, GHCR releases     | [docs/development/ci.md](docs/development/ci.md)                           |
| Argo CD GitOps (staging/production) | [docs/development/deploy-gitops.md](docs/development/deploy-gitops.md)     |

## Questions

Open a **Feature**, **Bug**, or **Chore** issue. For multi-step work, use one Feature issue with a checklist or split into several linked issues — see [docs/development/github-workflow.md](docs/development/github-workflow.md).
