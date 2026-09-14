## Why

Inventory reservations and catalog documents need independent runtimes. Listing creation also needs a durable event that inventory and the future Meilisearch projector can consume independently. `ProductCreated` replaces the synchronous web stock-seeding saga; a persisted listing can remain pending while consumers catch up.

## What Changes

- Add `services/inventory` with its own `inventory_db`, gRPC stock reads and ReserveStock, reservation Kafka consumer, inbox, and outbox. Remove reservations and live stock from products.
- Products atomically persists the catalog listing and a `ProductCreated` outbox event on `products.created`. The event carries catalog fields, explicit non-negative initial quantity, event identity, schema version, product version, and occurrence time; Kafka records are keyed by product id.
- Inventory consumes ProductCreated in its own consumer group and seeds stock transactionally and idempotently. EnsureStock is an internal operation, not a gRPC API.
- Web creates the listing with the requested initial quantity and reports creation with availability pending. Remove web EnsureStock and compensating deletion. Products never calls inventory.
- Catalog reads have no live stock join. Web may GetStock on the PDP until the projected read model exists; missing stock is pending and read failures are unavailable, never invented zero stock.
- Add-to-cart stamps the PDP snapshot. Checkout re-batches catalog prices and synchronously calls inventory ReserveStock.
- The future Meilisearch projector consumes the same catalog event independently, plus InventoryUpdated for availability. Implementing Meilisearch and catalog update/delete events remains #7.
- Retarget inventory CDC to inventory_db; add the products outbox connector/topic, Helm/runtime configuration, GHCR images, Cilium policy, secrets, and CI enrollment.
- Shop-dev data may be wiped; no backfill. Browse stays dark until #7.

## Capabilities

### New Capabilities

- `inventory`: Independent stock ledger, idempotent ProductCreated initialization, reservations, stock reads, and reservation Kafka.

### Modified Capabilities

- `products`: Catalog-only reads and durable ProductCreated publication; no inventory RPC or live quantity ownership.
- `web`: Asynchronous listing readiness, snapshot stamping, and synchronous checkout reserve.
- `argocd-gitops`: Inventory workload plus products creation-event transport.
- `cilium-mesh-policy`: Web stock reads/reserve only; no products-to-inventory access.
- `external-secrets`: Inventory database credentials through Doppler/ESO.
- `github-actions-ci`: Inventory lint, path filters, tests, and vulnerability scans.

## Impact

Use the current SQL catalog plus a transactional products outbox and Debezium now. The Mongo catalog cutover (#57) must preserve atomic listing/event persistence with an equivalent durable mechanism. Requested initial quantity is creation intent, not live catalog stock. Meilisearch (#7) remains a future consumer; this change establishes its catalog event source without building the projector.

Architecture and migration decisions are in `docs/catalog-inventory-search.md`. This supersedes the synchronous EnsureStock/compensating-delete plan and the earlier decision against ProductCreated. Reservation event contracts remain compatible while ownership moves to inventory.
