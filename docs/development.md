# Development

## Kubernetes Development (Current)

- Orchestration:
  - `Tiltfile` uses Helm chart under `infra/charts/refurbished-marketplace`
  - development values are in `infra/development/k8s/dev-helm-values.yaml`
  - Helm release namespace is `ecommerce`
  - CloudNativePG operator installed through Helm in namespace `cnpg-system`
- Secrets:
  - secrets are applied separately via `infra/development/k8s/secrets.yaml`
  - service and migration manifests consume secrets using `secretKeyRef`
- Namespaces:
  - application resources (`web`, `users`, `products`, `orders`, db clusters, migration jobs) are deployed to `ecommerce`
  - CloudNativePG operator stays in `cnpg-system`
- Database readiness:
  - service deployments use `initContainer` with `pg_isready` before app container startup
- Migration jobs:
  - Helm hook jobs (`pre-install,pre-upgrade`) in `templates/migrations.tpl`
  - users migration enabled by default
  - users migrator image built from `infra/development/docker/users-migrator.Dockerfile` (base: `ghcr.io/kukymbr/goose-docker:3.27.0`)
  - products migration enabled by default
  - products migrator image built from `infra/development/docker/products-migrator.Dockerfile` (base: `ghcr.io/kukymbr/goose-docker:3.27.0`)
- Service ports:
  - web: `8080`
  - users gRPC: `9091`
  - products gRPC: `9092`
  - orders gRPC: `9093`
  - users-db: 5432 (Postgres)
  - products-db: 5433 (Postgres)
  - orders-db: 5434 (Postgres)

## Service Discovery (Current)

- Web service upstream addresses are defined in values under `services.web.env`.
- Current values include:
  - `USERS_SVC_ADDR=users:9091`
  - `PRODUCTS_SVC_ADDR=products:9092`
  - `ORDERS_SVC_ADDR=orders:9093`

## Environment and Tooling

- Nix + direnv enabled.
- `.envrc` supports Colima/Testcontainers compatibility.
- `Makefile` includes:
  - `generate-proto`
