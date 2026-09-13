# Catalog, Inventory, and Search

Target architecture for `extract-inventory-service`: products durably emits ProductCreated; inventory and the future Meilisearch projector consume it independently. The implementation replaces the synchronous EnsureStock web flow; see the change tasks for deployment verification status. Meilisearch remains #7, and Mongo catalog cutover remains #57.

## Ownership

- Products owns listing identity, name, description, price, and merchant. SQL is the current store. Requested initial quantity is creation intent carried in the event, not live catalog stock.
- Inventory Postgres owns available/reserved quantity, reservations, seed intent, inbox, and outbox. Web calls stock reads and ReserveStock; products never calls inventory.
- Cart Redis owns displayed product snapshots and requested cart quantities.
- Meilisearch is a future catalog/availability projection, never the checkout source of truth.

## Creation and durable publication

Web sends catalog fields and explicit non-negative initial quantity to CreateProduct. Missing quantity is invalid; explicit zero is valid. Products commits the listing and a products outbox record atomically. Debezium publishes ProductCreated to `products.created`, keyed by product id. Never save the listing and then rely on a separate best-effort Kafka publish.

The event contains a stable event id, schema version, occurrence time, product id, initial product version, catalog fields, and explicit initial quantity. Outbox retries retain the same event identity. When catalog moves to Mongo, listing/event persistence must retain an equivalent atomic durability guarantee.

Inventory consumes the event in its own group, validates it, and commits inbox identity and initial stock in one transaction before acknowledging it. EnsureStock is internal inventory logic, not an RPC. Duplicate/replayed creation must never overwrite stock changed by reservations; conflicting seed intent is rejected. Transient failures remain retryable and do not delete the listing.

```mermaid
flowchart LR
  W[Web] -->|CreateProduct with initial quantity| P[Products]
  P -->|One transaction| C[(Catalog and products outbox)]
  C -->|CDC: ProductCreated| K[Kafka products.created]
  K -->|Inventory consumer group| I[Inventory]
  I -->|One transaction| L[(Stock and inventory inbox)]
  K -.->|Future projector consumer group| S[Search projector]
  I -.->|Future InventoryUpdated via outbox| S
  S -.-> M[(Meilisearch / PDP projection)]
  W -->|Synchronous ReserveStock| I
```

```mermaid
sequenceDiagram
  participant B as Browser
  participant W as Web
  participant P as Products
  participant D as Catalog and outbox
  participant K as Kafka
  participant I as Inventory
  participant S as Future search projector
  B->>W: Create listing + explicit initial quantity
  W->>P: CreateProduct
  P->>D: Commit listing + ProductCreated atomically
  D-->>P: Committed
  P-->>W: Listing identity
  W-->>B: Listing created; availability processing
  D-->>K: CDC publishes ProductCreated
  par Inventory group
    K->>I: ProductCreated
    I->>I: Commit inbox + stock seed
  and Future projector group
    K->>S: ProductCreated
    S->>S: Index catalog fields; availability not yet assumed
  end
```

Broker or consumer downtime may delay stock readiness, but committed listing creation succeeds. Failed catalog/outbox persistence fails creation and rolls back both. There is no web stock-seeding call and no compensation delete.

## PDP, cart, and checkout

Until the projected read model exists, web may GetStock once on PDP. A missing row means initialization is pending, a transport failure means availability is unavailable, and an existing row with zero available quantity means out of stock. Do not collapse these into zero. Disable purchase controls while availability is pending or unavailable.

Add-to-cart copies the displayed name/price snapshot. Cart quantity changes reuse that snapshot and cart reads do not hydrate products or stock. Checkout re-batches products for authoritative prices and synchronously reserves inventory before creating hosted payment. A checkout that reaches inventory before initialization fails safely and does not open payment.

```mermaid
sequenceDiagram
  participant B as Browser
  participant W as Web
  participant P as Products
  participant I as Inventory
  participant C as Cart Redis
  B->>W: GET product detail
  W->>P: GetProductByID
  W->>I: GetStock (until projection exists)
  W-->>B: Catalog and ready/pending/unavailable stock state
  B->>W: Add to cart with displayed snapshot
  W->>C: AddCartItem name, price, quantity
  B->>W: Checkout selected merchant
  W->>P: GetProductsByIDs for current prices
  W->>I: ReserveStock for created order
  Note over W,I: Hosted payment only after successful reserve
```

## Meilisearch follow-up (#7)

The projector consumes ProductCreated in a separate consumer group so inventory lag/replay does not affect indexing progress. It also needs InventoryUpdated with stock quantities and an inventory version; existing inventory.reserved/reservation_failed outcomes do not supply that snapshot.

Catalog and inventory events can arrive in either order. Retain stock updates that arrive before catalog data and merge them when catalog arrives. Apply monotonic catalog and inventory versions independently; a catalog update must not erase newer stock. Catalog creation alone must not mark a product in stock.

Before enabling a mutable production index, add catalog update/delete events and tombstones. This extraction establishes the durable creation topic and inventory consumer; the search consumer, InventoryUpdated publisher/projection, and update/delete events remain follow-up work. Browse stays dark until #7.

## Deployment and verification

Inventory has its own CNPG cluster, ESO credentials, GHCR images, and Cilium policy allowing web stock reads/reserve. Its outbox connector targets inventory_db. A separate products outbox connector publishes products.created using catalog credentials and appropriate secret access. Preserve event identity and tracing metadata.

CI must enroll inventory in lint, path filters, tests, and vulnerability scans. Verification covers command-to-Kafka replay, single reservation event emission, payment failure/expiry release, all-or-nothing command failure, missing/batch reads, and creation-event retry/atomicity. Kafka failure tests must wait for committed processing evidence rather than merely observing unchanged initial stock.

Cut over web and event consumers together so old synchronous initialization does not compete with ProductCreated seeding. No backfill is planned; the approved talos-dev wipe and live transport readiness require recorded verification. The task checklist is in `openspec/changes/extract-inventory-service/tasks.md`.
