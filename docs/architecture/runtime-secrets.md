# Runtime Secrets and Credentials

## Current Approach (Development)

- Kubernetes secrets are defined in `infra/development/k8s/secrets.yaml`.
- Tilt applies `secrets.yaml` directly before Helm-rendered manifests.
- Helm templates reference secrets via `secretKeyRef` only.

## Why This Structure

- Keeps secrets out of chart values and templates.
- Makes later migration to External Secrets Operator straightforward.
- Keeps app manifests stable while secret source changes.

## Current Secret Names

- `users/users-app` (DB username/password)
- `users/users-auth` (`JWT_SECRET`)
- `products/products-app` (DB username/password)
- `orders/orders-app` (DB username/password)

## Environment Variables by Service

These env vars are for internal services and edge service runtime configuration; external REST remains at the web edge.

### users

- `DB_URL` (built from DB user/password secret refs + host/db name)
- `JWT_SECRET` (from `users-auth`)

### products

- `DB_URL` (built from DB user/password secret refs + host/db name)

### orders

- `DB_URL` (built from DB user/password secret refs + host/db name)

## Future External Secrets Operator Path

1. Replace `infra/development/k8s/secrets.yaml` with `ExternalSecret` resources.
2. Keep secret names/keys the same.
3. No deployment/service template changes needed.
