## Why

Issue [#56](https://github.com/phuchoang2603/refurbished-marketplace/issues/56) (epic [#55](https://github.com/phuchoang2603/refurbished-marketplace/issues/55)): catalog documents will live on MongoDB, but talos-dev has no Mongo replica set. Standalone mongod cannot do change streams or multi-document transactions that P1/P2 need. This change deploys **MongoDB Controllers for Kubernetes (MCK) Community** the same way CNPG/Strimzi land: operator + CRDs, then a database CR. The shop must not read or write Mongo when this lands.

## What Changes

- Add an Argo child for the MCK operator (upstream Helm `mongodb/mongodb-kubernetes`) under `infra/charts/operators/mongodb/`, namespace `operators`, sync wave `0`.
- Add an Argo child (wave `2`) that applies a `MongoDBCommunity` replica set in `ecommerce` (`members: 1` on talos-dev).
- Sync SCRAM user passwords from Doppler via External Secrets (no plaintext in Git).
- Allow `products` → Mongo `27017` with a Cilium identity allow-list (no SPIRE mTLS on mongod).
- Optionally set unused host-only `MONGO_ADDR` on the products Deployment.
- Document operator, CR, Doppler keys, and that Community MCK does not include Ops Manager backup.

## Capabilities

### New Capabilities

- `mongodb-catalog`: MCK Community replica set in `ecommerce` as the future catalog document store (GitOps deploy, replica-set health, secrets, reachability). No application catalog cutover in this change.

### Modified Capabilities

- `argocd-gitops`: App-of-apps SHALL include the MCK operator Application and the `ecommerce` MongoDBCommunity Application with documented sync waves and namespaces.
- `cilium-mesh-policy`: Documented allow-list for products → mongod on `27017` without Cilium mutual auth; unknown callers dropped when enforce is on.
- `external-secrets`: Doppler-backed ExternalSecrets SHALL provision Mongo user (and related) credentials consumed by the `MongoDBCommunity` CR.

## Impact

- New charts under `infra/charts/operators/mongodb/` and a small `ecommerce` chart (or templates) for the CR + CNP + ExternalSecret.
- `infra/argocd/app-of-apps/values.yaml` and `infra/argocd/prod/root.yaml` valueFiles as needed.
- `docs/deployment/gitops.md`, `docs/development/secrets.md`.
- Marketplace `values.yaml` products env placeholder only.
- Issue: implements [#56](https://github.com/phuchoang2603/refurbished-marketplace/issues/56).

Non-goals: catalog data migration (#57), Meilisearch (#7), Go Mongo driver beyond unused env, Atlas, Ops Manager / Cloud Manager / `spec.backup`, 3-member replica set on talos-dev, splitting products/inventory services.
