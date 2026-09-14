## Context

See proposal.md for motivation. Mongo listings, `catalog_outbox`, and Debezium `products.created` already exist. Inventory consumes that topic in `inventory-product-created`. Products has no Kafka consumer today. Public `/` and `/products` render an empty catalog on purpose. Seller `/seller/products` still calls `ListProducts` and filters by merchant in web. Meilisearch is not in GitOps.

## Goals / Non-Goals

**Goals:**

- Run Meilisearch in `ecommerce` with the same Doppler + CNP wrapper-chart shape as catalog Mongo.
- Project `ProductCreated` from a products-process consumer group into a versioned index; rebuild from Mongo listings.
- Replace `ListProducts` with `SearchProducts`; undark browse without showing stock on cards.

**Non-Goals:**

- A second projector Deployment or image.
- `InventoryUpdated`, live qty in documents, or qty on browse cards.
- Catalog update/delete events.
- Serving GetProductByID / checkout from Meilisearch.

## Decisions

### 1. Wrapper chart + app-of-apps child, wave 2

Add `infra/charts/meilisearch/` (upstream [meilisearch-kubernetes](https://github.com/meilisearch/meilisearch-kubernetes) via Chart.yaml dependency) and an `apps.meilisearch` child destining `ecommerce`, sync wave `2` (with catalog Mongo, after ESO). Do not template an `ecommerce` Namespace. Chart owns ExternalSecret, single-node PVC, and the Meili CNP. Master key is Doppler `MEILI_MASTER_KEY` (Meilisearch requires at least 16 bytes) injected as `MEILI_MASTER_KEY`; API clients send `Authorization: Bearer`.

**Rationale:** Same split as Mongo (data plane not stuffed into the marketplace chart). Wave 2 so the Secret and Service exist before marketplace products (later wave) starts.

**Alternatives considered:** Bit inside marketplace chart (couples search lifecycle to shop Deployments); Meili Cloud (out of epic); multi-node (out of scope).

### 2. Projector lives in the products process

Start a second goroutine with `runtime.StartKafkaConsumer` (same helper inventory uses) group `products-search-product-created`. Official `meilisearch-go` client. Upsert is keyed by listing UUID (`id`). Configure index settings before documents: searchable `name`, `description`; filterable `merchant_id`; sortable `created_at`. Do not index `initial_qty`. Idempotent AddDocuments is the inbox: no SQL table. Transient Meili/Kafka errors retry; they must not share inventory’s groups.

Rebuild: scan Mongo `listings` and AddDocuments (operator-triggered or products startup ensure, documented either way — prefer an explicit rebuild path in products so talos-dev existing listings appear without replaying all Kafka).

**Rationale:** Issue #7 allowed products vs dedicated job. Products already owns catalog Mongo and will own `SearchProducts`. A separate image/CI matrix is unnecessary if Kafka consumption is isolated from ServeGRPC. Inventory isolation is the consumer group, not the binary.

**Alternatives considered:** Dedicated `search-projector` Deployment (cleaner blast radius, extra GHCR/Helm/CNP identity); Mongo change streams as the Meili path (proposal non-goal; Kafka is the shared bus).

### 3. SearchProducts replaces ListProducts in one proto cut

Add `SearchProducts` (query string, optional merchant_id, limit, offset). Remove `ListProducts` from proto and generated clients. Web catalog: empty query. Seller list: merchant filter = authenticated user id. No in-process merchant filter. PDP unchanged: `GetProductByID` + inventory `GetStock`.

Marketplace products env: `KAFKA_BOOTSTRAP_SERVERS`, `MEILI_URL`, master key from the Meili Secret. Optional `wait-for-meili` init hitting `/health` (unauthenticated).

**Rationale:** User asked not to keep ListProducts. Internal gRPC only; one PR can switch web and products together.

**Alternatives considered:** Keep ListProducts on Mongo for sellers (rejected); web still List-all then filter (rejected).

### 4. Meili CNP owned by the Meili chart

Ingress TCP/HTTP 7700 from `app: products` and kubelet host entities. No Cilium `authentication.mode: required`. Default-deny when `meshPolicy.enforce` is true. Do not teach marketplace `mesh-policy.tpl` Meili’s labels.

**Rationale:** Mirrors catalog Mongo CNP ownership.

## Risks / Trade-offs

- [Create-only index] → Listing edits/deletes will not update Meili until a later change. Acceptable while products has no update/delete RPCs.
- [Browse empty until projection/rebuild] → Rebuild from Mongo on first deploy so existing talos-dev listings appear; document lag.
- [Products restart if projector panics] → Treat consumer errors like inventory: log and retry in the helper; do not Fatal the process on a single Meili 5xx.
- [Master key rotation] → ESO refresh + Meili/products restart; document with other Doppler keys.
- [Helm-only commit ImagePullBackOff] → Products image must ship in the same rollout that removes ListProducts.

## Migration Plan

1. Add Doppler `MEILI_MASTER_KEY` in `dev`/`prd`.
2. Sync Meili Application; confirm Ready and CNP.
3. Ship products image with consumer + SearchProducts; switch web; delete ListProducts.
4. Rebuild index from Mongo; create a listing and confirm browse.
5. Rollback: revert the marketplace/web/products revision. Meili can stay deployed unused.

## Open Questions

None that change specs or task shape. Exact upstream chart version is pinned at apply time in Chart.lock.
