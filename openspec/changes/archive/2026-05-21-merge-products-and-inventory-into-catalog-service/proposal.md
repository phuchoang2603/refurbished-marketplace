## Why

The current split between products and inventory adds network and workflow complexity to basic marketplace behavior such as browsing stock and creating seller listings. This change reunifies catalog, stock, and reservation persistence into a single catalog service so reads and writes match the way the marketplace actually uses product data.

## What Changes

- **BREAKING** Consolidate the standalone `products` and `inventory` runtimes into one catalog service runtime with one database boundary.
- Move product records, inventory records, and reservation records under the same service ownership so listing creation and stock management can happen without cross-service orchestration.
- Replace stock lookups that currently require separate products and inventory service calls with a unified catalog read path, while keeping exact stock exposure limited to detail and admin-oriented surfaces in the first phase.
- Require initial stock explicitly in the first unified create path instead of defaulting omitted stock to zero.
- Keep reservation logic as a distinct module/table set inside the catalog service instead of removing it or flattening it into product rows.
- Do not introduce `product.created` plus outbox-driven inventory bootstrap for this domain split.
- Non-goal: moving the catalog source of truth to Elasticsearch or another search-first data store.
- Non-goal: changing buyer order semantics, payment flow semantics, or introducing a separate search indexing architecture in this change.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `products`: product requirements change from catalog-only ownership to unified catalog ownership that includes colocated stock data needed for listing reads and writes.
- `inventory`: inventory requirements change from a standalone service boundary to inventory behavior implemented inside the unified catalog service while preserving reservation semantics.

## Impact

- Affected code: `services/products`, internal gRPC callers that depend on reservation behavior, Kafka consumer wiring for reservations, service startup, schema baseline, and tests.
- APIs: inventory behavior is absorbed into the products/catalog boundary; web-facing migration is intentionally deferred from this first phase.
- Dependencies: the service boundary merge removes the standalone inventory runtime and its live contract surface.
- Systems: this change supersedes the need for async product-created inventory bootstrap and simplifies seller/admin product management flows.
