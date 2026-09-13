## Why

Epic [#55](https://github.com/phuchoang2603/refurbished-marketplace/issues/55) split **data** (Mongo listings vs Postgres stock) but kept one products binary. That collocation is the wrong runtime boundary now: reservations, Kafka inventory outbox, and catalog documents should scale and fail independently. This change extracts **inventory as its own service** and leaves products as catalog (Mongo in P1 of the epic). Shop-dev data MAY be wiped; browse stays dark until Meilisearch (#7).

## What Changes

- Add `services/inventory` (gRPC, own `inventory_db`, goose/sqlc, reservation Kafka consumer + `inventory_outbox`).
- Move `ReserveStock`, reservation records, and inventory qty out of the products process. **BREAKING** for callers that hit `products.v1.ReserveStock`.
- Products is catalog only. It SHALL NOT gRPC inventory. `CreateProduct` persists the listing and returns; web then `EnsureStock`. On inventory failure web compensating-deletes the listing.
- `GetProductByID` / `GetProductsByIDs` return catalog fields only (no live qty join). PDP qty comes from a later read model (`InventoryUpdated` → Mongo/Meili) or an optional **web→inventory** GetStock until that projector exists.
- Add-to-cart copies name/price from the page snapshot; no second GetProduct. Checkout re-batch products for price SoR and calls inventory `ReserveStock`.
- Drop `inventory.product_id` FK to `products`. Wipe talos-dev catalog/inventory rather than backfill.
- `ListProducts` is not replaced here; storefront list stays dark until #7.
- Helm, GHCR, Cilium, Doppler/ESO, CNPG cluster, and mesh policy for the new Deployment.

## Capabilities

### New Capabilities

- `inventory`: Stock ledger, reservations, EnsureStock, ReserveStock, Kafka reservation path, own Postgres. No catalog documents.

### Modified Capabilities

- `products`: Catalog only; no ReserveStock; no inventory RPC; Get* has no stock join; compensating delete for failed EnsureStock.
- `web`: Orchestrates create (products then inventory) and checkout reserve; stamps cart from the PDP snapshot; may GetStock for PDP until the read model exists.
- `argocd-gitops`: Marketplace chart deploys inventory alongside products.
- `cilium-mesh-policy`: Documented identities for web→inventory only (not products→inventory).
- `external-secrets`: Doppler keys for `inventory_db` / inventory app secret.

## Impact

- New proto `shared/proto/inventory/v1`, products proto drops reservation RPCs.
- `services/products` loses inventory SQL, Kafka consumer, ReserveStock.
- `services/web` clients and Cilium CNPs.
- CNPG: new Cluster; products_db migrations drop `products` FK and move/drop inventory tables (wipe).
- CI module tests, GHCR image, `docs/` ownership notes.
- Architecture: `docs/catalog-inventory-search.md`. Does **not** implement Meilisearch. Does **not** require Mongo outbox for stock.
- Reverses “no inventory microservice” in epic #55 / archived merge-catalog-service notes; update those issues/docs when applying.
