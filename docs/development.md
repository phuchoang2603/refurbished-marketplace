# Development

## Kubernetes Development (Current)

- Orchestration:
  - `Tiltfile` uses Helm chart under `infra/charts/refurbished-marketplace`
  - development values are in `infra/charts/refurbished-marketplace/values.yaml`
  - Helm release namespace is `ecommerce`
  - CloudNativePG operator installed through Helm in namespace `cnpg-system`
- Secrets:
  - secrets are applied via `infra/k8s/secrets.yaml` (Tilt applies it)
  - service and migration manifests consume secrets using `secretKeyRef`
- Namespaces:
  - application resources (`web`, `users`, `products`, `orders`, db clusters, migration jobs) are deployed to `ecommerce`
  - CloudNativePG operator stays in `cnpg-system`
- Database readiness:
  - service deployments use `initContainer` with `pg_isready` before app container startup
- Migration jobs:
  - Helm hook jobs (`pre-install,pre-upgrade`) in `templates/migrations.tpl`
  - users migrator (and other services) image built from `infra/docker/users-migrator.Dockerfile` (base: `ghcr.io/kukymbr/goose-docker:3.27.0`)
- Service ports:
  - web: `8080`
  - users gRPC: `9091`
  - products gRPC: `9092`
  - orders gRPC: `9093`
  - cart gRPC: `9094`
  - inventory gRPC: `9095`
  - users-db: port-forward `5432 -> 5432` (in-cluster Postgres is `5432`)
  - products-db: port-forward `5433 -> 5432` (in-cluster Postgres is `5432`)
  - orders-db: port-forward `5434 -> 5432` (in-cluster Postgres is `5432`)
  - inventory-db: port-forward `5435 -> 5432` (in-cluster Postgres is `5432`)

## Service Discovery (Current)

- Web service upstream addresses are defined in values under `services.web.env`.
- Current values include:
  - `USERS_SVC_ADDR=users:9091`
  - `PRODUCTS_SVC_ADDR=products:9092`
  - `ORDERS_SVC_ADDR=orders:9093`
  - `CART_SVC_ADDR=cart:9094`

## Environment and Tooling

- Nix + direnv enabled.
- No `.envrc` is committed to the repo; use your local direnv setup as preferred.
- `Makefile` includes:
  - `generate-proto`

## Prereqs (typical)

- `kubectl`, `helm`, `tilt`
- A Kubernetes cluster (for example: kind, Docker Desktop, Colima+k3s, etc.)
- Docker-compatible container runtime for building images
