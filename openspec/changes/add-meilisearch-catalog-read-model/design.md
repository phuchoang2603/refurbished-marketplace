## Context

Today `CreateProduct` writes `products` + `inventory` in one Postgres transaction and emits **no** catalog event. `ListProducts` is SQL `ORDER BY created_at LIMIT/OFFSET` with no text search or useful filters. Inventory already uses `inventory_outbox` → Debezium → Kafka; that is the template for catalog projection.

Issue #7 originally named Elasticsearch. The chosen engine is **Meilisearch**: typo-tolerant search, filters/facets, official Helm chart, no operator/CRD lifecycle — appropriate for Colima/staging beside Kafka/Istio/CNPG.

Delivery is phased so GitOps deploy can land before the CQRS cutover (#7 Phase 0 → 1 → 2).

## Goals / Non-Goals

**Goals:**

- PostgreSQL remains source of truth for catalog and stock writes.
- Meilisearch is the storefront catalog **read model** for list/search.
- Reliable projection via outbox + Debezium + projector (same family as inventory).
- GitOps-native Meilisearch deploy as an **app-of-apps** Helm Application (local + staging Argo), not a marketplace sidecar and not a marketplace-chart companion.
- Clear rebuild path when the index drifts.

**Non-Goals:**

- Elasticsearch / Typesense / ClickHouse for this read model.
- Multi-node Meilisearch HA or Cloud.
- Replacing detail/`GetProductByID` stock-aware reads (those stay on Postgres).
- Perfect real-time stock on every list row in v1 if we defer qty projection (see decisions).

## Decisions

### 1. Meilisearch over Elasticsearch / Typesense

Use the official `meilisearch/meilisearch` Helm chart as a single-node PVC-backed instance.

**Rationale:** Lightweight ops, MIT-friendly engine, good enough relevance/filters for a refurbished catalog; avoids ECK weight and Typesense HA complexity we do not need yet.

**Alternatives considered:** Elasticsearch (issue original — too heavy); Typesense (stronger OSS HA/curation — revisit if merchandising/HA becomes a requirement); ClickHouse (analytics-shaped, weaker storefront search DX).

### 2. App-of-apps Meilisearch chart — not sidecar, not marketplace companion

Deploy Meilisearch via a dedicated chart under `infra/charts/meilisearch/` (upstream official Helm chart) registered in `infra/argocd/app-of-apps` for local and staging. Products reaches it over stable in-cluster DNS. Do **not** run Meilisearch as a products pod sidecar; do **not** fold it into the marketplace chart as a companion Deployment.

**Rationale:**

- Catalog index is a **shared durable read model** (PVC): per-pod sidecars diverge and reindex-on-restart. Cart Valkey stays a localhost sidecar for **ephemeral** cache only (`services.<name>.sidecars` in the marketplace chart).
- App-of-apps matches Kafka/observability-style platform services: own lifecycle, sync wave, enable/disable without rewriting marketplace service Deployments.
- Keeps marketplace chart focused on ecommerce services; Meilisearch has different scaling/backup story than cart sidecar.

**Alternatives considered:**

| Option                                 | Why not                                                                                                                                       |
| -------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Products sidecar (cart pattern)        | Divergent indexes; no shared PVC; broken projector under multi-replica                                                                        |
| Marketplace-chart companion Deployment | Couples Meili lifecycle to marketplace Tilt chart; blurs platform vs app boundary while app-of-apps already ships platform deps (Kafka, etc.) |
| Elasticsearch operator wave            | Rejected for ops weight                                                                                                                       |

**Related:** Universal marketplace `sidecars:` is reserved for Valkey-like cases; Meilisearch stays out of that abstraction.

### 3. Catalog outbox + Debezium + in-process projector

Add `products_outbox` mirroring `inventory_outbox`. `CreateProduct` inserts an outbox row in the same TX. Debezium publishes to Kafka. A projector inside the products service (or a dedicated binary if it grows) upserts/deletes Meilisearch documents.

**Rationale:** Matches existing CDC patterns; keeps Postgres commit and “will be projected” coupled; retries via Kafka consumer semantics.

**Alternatives considered:** Sync HTTP index inside the write TX (couples availability); poll Postgres (laggy, no event reuse).

### 4. v1 document: catalog fields + `available_qty` from create; reservation updates deferred unless cheap

Index document includes at least: `product_id`, `name`, `description`, `price_cents`, `merchant_id`, `available_qty`, `created_at`. On create, set `available_qty` from initial stock. **Reservation commit/release projection into Meilisearch is a follow-up task** in Phase 1 if time allows; otherwise document lag and keep detail pages on Postgres for exact stock.

**Rationale:** Unblocks search/filters (merchant, in-stock at create time) without blocking on full reservation event fan-in; list API today already omits live stock.

**Alternatives considered:** Full qty projection in the same slice (more correct, more scope); catalog-only with no qty (weaker in-stock filter).

### 5. New `SearchProducts` gRPC; evolve storefront `ListProducts` to Meilisearch

Prefer an explicit `SearchProducts` (query string + filters + pagination). Storefront/web list paths call it (empty query = browse). Keep Postgres-backed paths only where needed for admin/debug if any; default list for web MUST NOT hit SQL offset scans.

**Rationale:** Clear CQRS boundary; avoids silently changing every internal `ListProducts` caller without intent.

**Alternatives considered:** Only change `ListProducts` in place (simpler proto, muddier semantics).

### 6. Secrets via External Secrets; dedicated namespace OK

`MEILI_MASTER_KEY` from ESO/Doppler. Prefer a dedicated namespace (e.g. `meilisearch`) or shared platform namespace consistent with other app-of-apps platform services. Products gets `MEILI_URL` + key via env/secretRef from marketplace values.

**Rationale:** Platform search is not application business logic; secrets and network policy can follow the same patterns as Kafka reachability into `ecommerce`.

## Risks / Trade-offs

| Risk                                                            | Mitigation                                                                         |
| --------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Eventual consistency: create succeeds but search misses briefly | Document lag; projector retries; UI tolerates refresh; rebuild job for drift       |
| Index drift after bugs / missed events                          | Versioned reindex from Postgres snapshot documented and scripted                   |
| Stale `available_qty` if reservation projection deferred        | Detail page stays Postgres; list filter is approximate until qty sync lands        |
| Meilisearch downtime blocks browse                              | Fail soft on web (unavailable page) like other deps; writes still work on Postgres |
| Colima RAM pressure                                             | Single-node, resource requests/limits; optional disable for non-search work        |
| Dual-write confusion                                            | Enforce: only projector writes Meilisearch; services never treat Meili as SoR      |

## Migration Plan

1. **Phase 0:** Ship Meilisearch Helm chart + app-of-apps Application + secrets; verify health from products network path. No catalog traffic yet.
2. **Phase 1:** Migrations + Debezium + projector + index settings; backfill/reindex existing products; create path projects forward.
3. **Phase 2:** Expose `SearchProducts` / wire web; flip storefront list off Postgres.
4. **Rollback:** Re-point list/search to Postgres queries; leave Meilisearch deployed idle; outbox/projector can pause without blocking writes.

## Open Questions

- Projector packaging: same `products` binary vs small `products-projector` deployable (default: same binary / loop beside Kafka consumers).
- Whether Phase 1 MUST include reservation → `available_qty` updates or explicitly defer to a follow-up issue.
- Meilisearch chart version pin, namespace, and resource limits for Colima vs staging.
- Local DX: rely on Argo-managed Meili after `tilt up` installs Argo (same as Kafka), optional port-forward only for debugging.
