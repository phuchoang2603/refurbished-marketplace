# Local setup

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A local Kubernetes runtime used by Tilt (for example Colima + Docker)
- [Tilt](https://tilt.dev/) (provided inside the devenv shell)
- [Doppler](https://www.doppler.com/) account for cluster secrets — see [secrets.md](secrets.md)

## Development shell

Enter the shell before generators, tests, or infrastructure commands:

```bash
devenv shell
```

The shell provides Go, protobuf tooling, database migration/query generators, Kubernetes tooling, Tilt, Doppler, and OpenSpec. A gitignored `.env` file is loaded automatically (`dotenv.enable` in `devenv.nix`).

## Tilt

After [secrets setup](secrets.md), start the local stack:

```bash
tilt up
```

Tilt uses the root `Tiltfile` to build services, apply Kubernetes/Helm resources under `infra/`, and keep the cluster in sync while you edit code. Use the Tilt UI for service status, logs, resource readiness, and rebuilds.

## Integration testing

Integration tests rely on Testcontainers for Kafka, PostgreSQL, and Redis/Valkey. Prefer verifying full-service flows through Tilt; run targeted Go tests when they add meaningful coverage.
