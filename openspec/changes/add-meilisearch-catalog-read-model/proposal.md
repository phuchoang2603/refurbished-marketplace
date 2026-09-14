## Why

Issue [#7](https://github.com/phuchoang2603/refurbished-marketplace/issues/7) (epic [#55](https://github.com/phuchoang2603/refurbished-marketplace/issues/55)): public browse is still an empty page on purpose. Mongo is the catalog write store and `products.created` already exists for inventory. Search and storefront lists do not belong on Mongo.

## What Changes

- Deploy Meilisearch in `ecommerce` via GitOps (wrapper chart, app-of-apps child, Doppler master key, Cilium policy for the projector identity).
- Products consumes `products.created` in a consumer group independent of inventory, upserts catalog fields into Meilisearch, and can rebuild the index from Mongo. Documents carry listing identity, name, description, price, merchant, and created time — not live stock and not `initial_qty` as availability.
- Add gRPC `SearchProducts` (text, optional merchant filter, offset/limit). Empty query is browse.
- **BREAKING (internal gRPC):** remove `ListProducts`. Web public catalog and seller list call `SearchProducts`. `GetProductByID` / `GetProductsByIDs` stay on Mongo. Product detail (PDP) still loads qty from inventory `GetStock`.
- Fail search/browse the same way other products read failures fail (localized unavailable page), without stalling inventory consumers if Meili is down.

## Capabilities

### New Capabilities

- `meilisearch-catalog`: GitOps-managed Meilisearch in `ecommerce` is the storefront catalog projection, not the listing or stock source of truth.

### Modified Capabilities

- `products`: SearchProducts from Meilisearch; projector on `products.created`; ListProducts removed; keyed catalog reads stay on Mongo.
- `web`: Public catalog and seller list use SearchProducts; browse shows catalog fields only; PDP qty stays on inventory.
- `argocd-gitops`: App-of-apps deploys Meilisearch into `ecommerce`.
- `external-secrets`: Doppler-backed Meilisearch master key; no plaintext key in Git.
- `cilium-mesh-policy`: Products may reach Meilisearch HTTP without treating it as a SPIRE hop; unknown callers denied.
- `mongodb-catalog`: Shop list/browse no longer reads listings via Mongo ListProducts; replica set remains create/get/batch SoT.
- `github-actions-ci`: Products tests fan out for Meilisearch test helpers if added.

## Impact

- `infra/charts/meilisearch/` (or equivalent wrapper), `infra/argocd/app-of-apps`, marketplace products env, Mongo CNP/Meili CNP, Doppler `MEILI_MASTER_KEY`, `shared/proto/products`, `services/products`, `services/web`, `docs/catalog-inventory-search.md`, secrets/GitOps/Cilium docs.
- Inventory, cart, checkout, and `GetStock` / `ReserveStock` stay as they are.
- Non-goals: `InventoryUpdated` and live qty in the index or on browse cards; catalog update/delete events and tombstones; serving PDP/cart/checkout from Meilisearch; Meili Cloud / multi-node HA; a separate projector Deployment.
