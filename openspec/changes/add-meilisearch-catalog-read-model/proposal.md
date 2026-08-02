## Why

Storefront catalog listing is still a Postgres `ORDER BY created_at LIMIT/OFFSET` query. That is fine for a tiny catalog but does not support typo-tolerant search or useful filters (merchant, in-stock, price). Issue #7 introduces CQRS for products: PostgreSQL stays the write model; a lightweight search engine becomes the read model. Meilisearch replaces the originally proposed Elasticsearch so local/staging ops stay small (official Helm chart, no operator/CRD wave) beside existing Kafka/Istio/CNPG workloads.

## What Changes

- Deploy **Meilisearch** via Helm under `infra/` and register it in the Argo CD app-of-apps (local + staging)
- Add a **catalog outbox** (`products_outbox`) and Debezium → Kafka path so `CreateProduct` projects into Meilisearch
- Add a **projector** that upserts/deletes Meilisearch documents from catalog events
- Version Meilisearch **index settings** in-repo (searchable/filterable attributes)
- Serve catalog **list/search** from Meilisearch (`SearchProducts` and/or evolved `ListProducts`); Postgres remains authoritative for writes and stock mutations
- Point **web** catalog/seller list flows at the new APIs with server-side filters
- Document eventual consistency, failure handling, and Postgres → Meilisearch **reindex/rebuild**

### Non-goals

- Elasticsearch, OpenSearch, ClickHouse, or Typesense as the catalog read model
- Replacing PostgreSQL as source of truth for products/inventory
- Multi-node Meilisearch HA / Meilisearch Cloud (single-node self-hosted for local + staging v1)
- Advanced merchandising (A/B ranking, synonyms packs) beyond basic ranking rules
- Making Meilisearch required for every local microservice iteration (optional Tilt port-forward only when developing search)

## Capabilities

### New Capabilities

- `catalog-search`: Meilisearch-backed catalog read model — index document shape, projection/sync, search/list query semantics, reindex, and consistency expectations

### Modified Capabilities

- `products`: `CreateProduct` MUST emit a catalog outbox event; list/search reads that are storefront-oriented MUST be served from the catalog search read model
- `argocd-gitops`: App-of-apps MUST include a Meilisearch Application (Helm chart + secrets wiring)
- `web`: Public catalog and seller product list MUST use the Meilisearch-backed products search/list APIs (no client-side merchant filtering when the API can scope it)

## Impact

- **Add:** `infra/charts/meilisearch/` (or equivalent Helm wrapper), app-of-apps entry, `products_outbox` migration/SQLC, Debezium entity, projector, Meilisearch client usage in products, index settings, docs
- **Update:** `services/products` (create outbox, search/list gRPC), `shared/proto/products`, `services/web` catalog handlers, `infra/charts/kafka` Debezium config, `infra/argocd/app-of-apps`
- **Issue:** Closes / implements [#7](https://github.com/phuchoang2603/refurbished-marketplace/issues/7) in phases (deploy → project → read/web)
