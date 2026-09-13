## Context

See proposal.md. Today `services/products` owns SQL `products` + `inventory` + reservations + Kafka in one process. This change splits the inventory **runtime**. Products never opens a gRPC client to inventory. Web is the BFF for create and reserve. Shop-dev MAY be wiped. Browse stays dark until #7. Target read path: inventory events update a product-page/search projection (Mongo and/or Meilisearch). Durable picture: `docs/catalog-inventory-search.md`.

## Goals / Non-Goals

**Goals:** Own CNPG `inventory` cluster; inventory gRPC to **web**; move reservation Kafka; web orchestrates listing + EnsureStock with compensating listing delete; cart stamp from PDP; Cilium + Helm + Doppler; drop catalog FK.

**Non-Goals:** Products→inventory gRPC; implementing Meilisearch in this change; Mongo outbox for stock; inventory owning listings; backfill; ListProducts replacement.

## Decisions

### 1. Web orchestrates create; products never dials inventory

Web: `CreateProduct` (catalog) then `EnsureStock`. On EnsureStock failure, web deletes the listing. Kafka is for reservation events and later `InventoryUpdated` read-model sync, not planting initial stock.

**Rationale:** No products→inventory coupling. Seller still waits for both stores before 200. BFF owns the two-step saga.

**Alternatives considered:** products EnsureStock RPC (rejected); `ProductCreated` Kafka (stock lag on create).

### 2. Catalog reads have no live stock join

`Product` qty fields MAY be empty on Get*. PDP uses a projected view when the read model lands. Until then web MAY call inventory GetStock for the detail page only. Add-to-cart copies name/price from the HTML/form; cart GET is Redis. Checkout uses products batch for **price** SoR and inventory ReserveStock for **qty**.

**Rationale:** Browse/cart scale without inventory QPS. Reserve stays strong.

**Alternatives considered:** products joins inventory on every Get* (rejected).

### 3. New `inventory_db` Cluster; wipe; no FK

Move inventory/reservation/outbox/inbox migrations under `services/inventory`. `product_id` UUID PK without REFERENCES products. Truncate/drop old tables on products_db. No backfill.

**Rationale:** User chose wipe on talos-dev. FK to a dying catalog table is the landmine.

### 4. New proto `inventory.v1`; remove ReserveStock from products

Generate with existing `generate-proto`. Debezium connectors that pointed at products_db inventory_outbox retarget inventory_db. Products grows a delete-listing RPC if missing, for create compensation.

**Rationale:** Event names (`inventory.reserved`) stay; publisher identity changes.

### 5. Enroll inventory in mesh; only web is a gRPC caller

Same CNP mTLS family as other marketplace gRPC. Allow `web` → inventory. Do not allow `products` → inventory.

### 6. Later: InventoryUpdated → Mongo/Meili (not this apply)

Projector (in products or a small worker) consumes inventory qty events and updates the search/PDP view. Checkout never trusts that view.

## Risks / Trade-offs

| Risk                              | Mitigation                                                      |
| --------------------------------- | --------------------------------------------------------------- |
| Create spans two systems from web | Compensate-delete listing; EnsureStock idempotent on product_id |
| PDP without projection has no qty | Optional web GetStock on detail only; then event-sourced view   |
| Spoofed add-to-cart price         | Checkout re-batch products for unit price                       |
| Dark browse                       | Accepted until #7                                               |
| Debezium slot on wrong DB         | Kafka chart connectors follow inventory_db                      |
| Epic #55 text says no split       | Update issue when applying                                      |

## Migration Plan

1. Land inventory service empty schema + Helm (wipe).
2. Cut web ReserveStock, EnsureStock, Kafka consumer; products drop inventory SQL.
3. Cart add stamps from the form; remove GetProduct on add.
4. Drop `products` table / FK when Mongo catalog cutover (#57) applies.
5. #7: Meili + InventoryUpdated projection for PDP/search.
6. Rollback: restore products binary that still has inventory (not dual-running writers).

## Open Questions

None. gRPC `:9097`. Mongo catalog cutover stays #57. Meili stays #7.
