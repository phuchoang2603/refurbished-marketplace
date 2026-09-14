## Why

Issue [#7](https://github.com/phuchoang2603/refurbished-marketplace/issues/7) (epic [#55](https://github.com/phuchoang2603/refurbished-marketplace/issues/55)): public browse is still an empty page on purpose. Mongo is the catalog write store and `products.created` already exists for inventory. Search and storefront lists do not belong on the write service.

## What Changes

- Deploy Meilisearch in `ecommerce` via GitOps (wrapper chart, app-of-apps child, Doppler master key, Cilium policy for the search identity).
- Add a **search** service (own binary, GHCR image, Helm workload). It consumes `products.created` in a group independent of inventory, upserts catalog fields into Meilisearch, and serves gRPC `SearchProducts`. Documents carry listing identity, name, description, price, merchant, and created time — not live stock and not `initial_qty` as availability. No Mongo rebuild: talos-dev data may be wiped; the index fills from new creates.
- **BREAKING (internal gRPC):** remove `ListProducts` from products. Web public catalog and seller list call search `SearchProducts`. Products keeps Create / GetProductByID / GetProductsByIDs on Mongo. Product detail (PDP) still loads qty from inventory `GetStock`.
- Search or Meili unavailability uses a localized catalog unavailable page. Products create/get and inventory consumers MUST NOT stall if Meili is down.

## Capabilities

### New Capabilities

- `meilisearch-catalog`: GitOps-managed Meilisearch in `ecommerce` is the storefront catalog projection, not the listing or stock source of truth.
- `search`: Read-side catalog service: ProductCreated projector plus SearchProducts. It does not own Mongo listings or stock.

### Modified Capabilities

- `products`: ListProducts removed; keyed catalog reads and creates stay on Mongo; products does not query Meilisearch or consume `products.created`.
- `web`: Public catalog and seller list use search SearchProducts; browse shows catalog fields only; PDP qty stays on inventory.
- `argocd-gitops`: App-of-apps deploys Meilisearch; marketplace chart deploys the search workload.
- `external-secrets`: Doppler-backed Meilisearch master key used by Meilisearch and the search service; no plaintext key in Git.
- `cilium-mesh-policy`: Search may reach Meilisearch HTTP without SPIRE; web may reach search gRPC; unknown callers denied.
- `mongodb-catalog`: Shop list/browse no longer reads listings via Mongo ListProducts; replica set remains create/get/batch SoT.
- `github-actions-ci`: Search module in lint/test/govulncheck matrices; Meilisearch test helpers fan out to search.
- `ghcr-release`: Search image is part of the release matrix.

## Impact

- `services/search`, `infra/docker` search image, `infra/charts/meilisearch/`, app-of-apps, marketplace search + web clients, Meili CNP, Doppler `MEILI_MASTER_KEY`, proto (`search` and products ListProducts removal), `docs/catalog-inventory-search.md`, secrets/GitOps/Cilium docs.
- Inventory, cart, checkout, products Mongo SoT, and `GetStock` / `ReserveStock` stay as they are.
- Non-goals: Mongo→Meili rebuild/backfill; `InventoryUpdated` and live qty in the index or on browse cards; catalog update/delete events and tombstones; serving PDP/cart/checkout from Meilisearch; Meili Cloud / multi-node HA; projector-only Job plus SearchProducts still on products.
