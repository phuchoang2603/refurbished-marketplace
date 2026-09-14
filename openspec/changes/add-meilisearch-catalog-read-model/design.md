## Context

See proposal.md for motivation. Mongo listings, `catalog_outbox`, and Debezium `products.created` already exist. Inventory consumes that topic in `inventory-product-created`. Products has no Kafka consumer and should stay the write model. Public `/` and `/products` render an empty catalog on purpose. Seller `/seller/products` still calls `ListProducts` and filters by merchant in web. Meilisearch is not in GitOps. Talos-dev catalog data will be wiped; there is no rebuild-from-Mongo path.

## Goals / Non-Goals

**Goals:**

- Run Meilisearch in `ecommerce` with the same Doppler + CNP wrapper-chart shape as catalog Mongo.
- Put the read model in its own binary: project `ProductCreated` and serve `SearchProducts`.
- Replace products `ListProducts` with search `SearchProducts`; undark browse without showing stock on cards.

**Non-Goals:**

- Projector inside the products process, or products proxying Meili.
- A projector-only Job plus a second search-API binary (one read-side service is enough).
- Rebuild/backfill from Mongo.
- `InventoryUpdated`, live qty in documents, or qty on browse cards.
- Catalog update/delete events.
- Serving GetProductByID / checkout from Meilisearch.

## Decisions

### 1. Wrapper chart + app-of-apps child, wave 2

Add `infra/charts/meilisearch/` (upstream [meilisearch-kubernetes](https://github.com/meilisearch/meilisearch-kubernetes) via Chart.yaml dependency) and an `apps.meilisearch` child destining `ecommerce`, sync wave `2` (with catalog Mongo, after ESO). Do not template an `ecommerce` Namespace. Chart owns ExternalSecret, single-node PVC, and the Meili CNP. Master key is Doppler `MEILI_MASTER_KEY` (Meilisearch requires at least 16 bytes) injected as `MEILI_MASTER_KEY`; API clients send `Authorization: Bearer`.

**Rationale:** Same split as Mongo. Wave 2 so the Secret and Service exist before marketplace search starts.

**Alternatives considered:** Bit inside marketplace chart; Meili Cloud; multi-node.

### 2. One read-side `search` binary (projector + query)

CQRS here is: products/Mongo is the command side; Meili is the query store; **search** is the query-side process.

```
CreateProduct ──▶ products / Mongo ──CDC──▶ products.created
                                            ├─ inventory-product-created
                                            └─ search-product-created ──▶ Meili
                                                                          ▲
Web browse / seller list ── SearchProducts ── search ─────────────────────┘
Web PDP / checkout ── GetProductByID / batch ── products / Mongo
```

`services/search`: `runtime.StartKafkaConsumer` group `search-product-created`, official `meilisearch-go`, upsert by listing UUID (`id`). Index settings before documents: searchable `name`, `description`; filterable `merchant_id`; sortable `created_at`. Do not index `initial_qty`. Idempotent AddDocuments is the inbox. gRPC `SearchProducts` (empty query = browse). No Mongo client. Fail search if Meili is down. Marketplace `app: search`, GHCR `search`, Kafka bootstrap + Meili URL/key, `wait-for-meili` on `/health`.

**Rationale:** The write service should not scale, restart, or fail with the read model. Inventory already shows the pattern: a consumer of `products.created` that is not products. Putting projector and SearchProducts in **one** read binary avoids a Job that can only write Meili while products still becomes a Meili proxy.

**Alternatives considered:** Projector in products (couples SoT to search); projector Job + SearchProducts on products (two hops, products still depends on Meili); web → Meili HTTP (breaks internal gRPC composition); two read binaries (projector vs API) — extra image/Helm/CNP for no current load reason.

### 3. SearchProducts on search; delete ListProducts on products

New `shared/proto/search/v1` (or equivalent) with SearchProducts (query, optional merchant_id, limit, offset). Remove `ListProducts` from products proto. Web catalog: empty query. Seller list: merchant filter = authenticated user id. PDP unchanged: products `GetProductByID` + inventory `GetStock`.

**Rationale:** List is a query-model concern. Keyed reads stay on the write model.

**Alternatives considered:** SearchProducts still on ProductsService implemented by search (confusing ownership); keep ListProducts on Mongo for sellers (rejected).

### 4. Meili CNP owned by the Meili chart; search is the only app caller

Ingress TCP 7700 from `app: search` and kubelet. No Cilium `authentication.mode: required`. Default-deny when enforce is true. Marketplace mesh: web → search gRPC with the same mutual-auth pattern as other shop gRPC services (chart already emits allow-<svc> for each grpc service).

**Rationale:** Mirrors catalog Mongo CNP ownership; products never dials 7700.

## Risks / Trade-offs

- [Create-only index] → Edits/deletes will not update Meili until a later change. Acceptable while products has no update/delete RPCs.
- [Wipe + no rebuild] → After cutover, browse fills only from new ProductCreated. Document that.
- [New service cost] → Extra module, image, CI matrix, Helm row — the point of isolating the read side.
- [Helm-only commit ImagePullBackOff] → Search (and products/web if proto-changed) images must ship with the chart revision.

## Migration Plan

1. Add Doppler `MEILI_MASTER_KEY` in `dev`/`prd`.
2. Sync Meili Application; confirm Ready and CNP.
3. Wipe talos-dev catalog/search data as planned.
4. Ship search + products (no ListProducts) + web; create a listing; confirm browse and seller list.
5. Rollback: revert marketplace/search/web/products revision. Meili can stay deployed unused.
