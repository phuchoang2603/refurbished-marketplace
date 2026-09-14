## Why

Issue [#57](https://github.com/phuchoang2603/refurbished-marketplace/issues/57) (epic [#55](https://github.com/phuchoang2603/refurbished-marketplace/issues/55)): listings still live in Postgres `products` after inventory extraction. Mongo is already in `ecommerce` and unused. Catalog documents belong on the replica set, while `ProductCreated` on `products.created` must stay as atomic as today’s SQL listing+outbox commit.

## What Changes

- Products reads and writes listing documents in Mongo (v1 fields match today’s catalog columns). `GetProductByID` / `GetProductsByIDs` / seller `ListProducts` use Mongo. Public browse stays dark until #7.
- CreateProduct commits the listing document and a catalog outbox document in one Mongo replica-set transaction. Debezium Mongo CDC (EventRouter) publishes the existing `ProductCreated` contract onto `products.created` with retained event identity. Dual-write (Mongo listing + SQL `products_outbox`, or Mongo listing + inventory insert) is out. Products does not also run an in-process Kafka publisher.
- Add `shared/testutil/mongo` (replica-set Testcontainers module) for products tests, enrolled in `go.work` and CI fan-out.
- Wire products to Mongo for real (driver, existing `catalog` SCRAM secret). `MONGO_ADDR`-only unused env is not enough.
- **BREAKING (talos-dev data):** drop SQL `products` / `products_outbox`, the Postgres products-outbox connector, and `products_db`. Wipe is acceptable; no backfill.
- Inventory and web checkout stay on the current Kafka/gRPC contracts. Meilisearch stays #7.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `products`: Catalog source of truth is Mongo; listing and ProductCreated persist atomically without Postgres; ListProducts is a Mongo find for seller list.
- `mongodb-catalog`: Shop catalog traffic depends on the replica set; products authenticates and uses it; Kafka Connect may read the catalog outbox.
- `argocd-gitops`: Marketplace no longer deploys products CNPG; Kafka keeps `products.created` and a Mongo outbox connector instead of the Postgres products connector.
- `external-secrets`: Products (and Connect) use Doppler-backed Mongo credentials; products Postgres app secret leaves with `products_db`.
- `cilium-mesh-policy`: Mongo 27017 allows products and Kafka Connect (no SPIRE); unknown callers still denied.
- `github-actions-ci`: `shared/testutil/mongo` path filters fan out to products.

## Impact

- `services/products` persistence, `shared/testutil/mongo`, Helm env, marketplace CNPG for products, `infra/docker/connect-debezium.Dockerfile`, kafka connector templates, Mongo CNP, Doppler keys, `docs/catalog-inventory-search.md`.
- gRPC `Product` / `ProductCreated` payloads stay compatible. Inventory code is unchanged if the topic contract holds.
- Non-goals: Meilisearch, `SearchProducts`, catalog update/delete events, `InventoryUpdated`, variants, moving stock to Mongo, in-process outbox relay, change streams as the Meili path.
