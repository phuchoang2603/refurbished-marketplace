## Context

This change separates catalog and inventory runtimes and uses ProductCreated to initialize inventory asynchronously. The current implementation contains a synchronous web EnsureStock saga; the remaining tasks replace it. Products uses SQL today; Mongo cutover remains #57. Meilisearch implementation remains #7.

## Goals / Non-Goals

**Goals:** Independent inventory database/runtime; durable catalog creation events; idempotent inventory initialization; pending availability in web; synchronous checkout reservation; deployment and CI enrollment.

**Non-Goals:** Products-to-inventory RPC, live stock in catalog, Meilisearch implementation, catalog update/delete event implementation, backfill, or replacing SQL catalog with Mongo in this change.

## Decisions

### 1. Persist listing and ProductCreated atomically

CreateProduct requires an explicit non-negative initial quantity. Products commits the listing and a products outbox row in one SQL transaction, then returns the listing identity. Debezium publishes the outbox to `products.created`; a separate best-effort publish after committing the listing is insufficient. Broker downtime delays publication without losing a committed event. Failure to persist either record rolls back both.

ProductCreated contains event_id, schema_version, occurred_at, product_id, product_version (initially 1), name, description, price_cents, merchant_id, and explicit initial_qty with presence. Kafka uses product_id as its key. Redelivery preserves the event identity and payload. Initial quantity is immutable creation intent, not a live catalog quantity. A future Mongo implementation must retain the same atomicity guarantee.

**Rationale:** Inventory and search can independently consume one durable catalog event without coupling seller creation to either consumer.

### 2. Inventory seeds through its own consumer

Inventory uses a consumer group separate from the future search projector. It validates ProductCreated, then records event_id in its inbox and seeds available_qty with reserved_qty zero in one transaction. EnsureStock remains an internal operation; remove its gRPC method and web client. Duplicate event delivery or replay for an already seeded product must not reset stock after reservations, commits, or releases. Store enough seed identity/intent to reject conflicting seeds without overwriting the ledger.

Transient failures roll back and remain retryable; acknowledge only after commit. Invalid events must not create stock or be treated as successfully seeded: expose the error for operational recovery using the repository's consumer error path. Consumer failure never triggers catalog deletion.

### 3. Creation succeeds before availability is ready

Web sends catalog details and explicit initial quantity to CreateProduct and reports that the listing was created while availability is processing. It does not call inventory or delete a listing to compensate consumer lag. Missing stock on PDP means pending initialization; transport errors mean unavailable. Neither becomes zero stock. A real row with zero available quantity means out of stock.

Until a projection exists, web may read GetStock once on PDP. Keep purchase controls disabled while availability is pending/unavailable. Checkout still uses catalog batch prices and synchronous inventory ReserveStock; a missing stock row fails reservation and prevents hosted payment. Add-to-cart uses the displayed snapshot; cart reads remain Redis-only.

### 4. Search is an independent future projection

In #7, a separate projector group consumes ProductCreated for catalog fields and InventoryUpdated for stock availability. Reservation outcome events alone do not contain a stock snapshot and cannot substitute for InventoryUpdated. Catalog arrival does not imply stock readiness. The projector must merge events in either arrival order, retain stock updates that arrive before catalog data, and apply monotonic versions separately for catalog and inventory. Catalog update/delete events and tombstones must be designed before enabling a mutable production search index. No projector or InventoryUpdated publisher is implemented in this extraction.

### 5. Preserve reservation semantics and runtime isolation

Inventory owns inventory/reservations/inbox/outbox tables without a catalog FK. It consumes orders.created and payment outcomes and exposes ReserveStock on :9097. Command-to-Kafka replay must not double-hold or emit a second successful reservation event. Full-order failure leaves no partial hold. Payment failure/expiry releases stock.

Web is the only application gRPC caller, with the same Cilium mutual authentication as other marketplace services. Products never calls inventory. Keep GetStock/GetStocksByIDs/ReserveStock and remove the temporary compensation-only DeleteProduct API unless a separate catalog deletion requirement is approved.

### 6. Deploy both event paths and enroll CI

Retarget inventory-outbox CDC to inventory_db. Add a separate products outbox connector, products.created topic, secret/RBAC wiring, and inventory subscription. Preserve CDC tracing metadata. Marketplace Helm deploys inventory, CNPG, ESO, and mesh policy with matching image tags. CI includes inventory in lint globs, path-filter outputs, tests, and vulnerability scans.

## Risks / Trade-offs

- Consumer lag delays purchasability: show pending availability and rely on strong checkout reservation.
- Broker/consumer outages: retain durable outbox events and retry without deleting listings.
- Duplicate events: transactional inbox plus product-level seed idempotency prevent stock reset.
- Cross-topic reordering: the future projector merges independent catalog and stock versions.
- Catalog migration: preserve listing/event atomicity when moving to Mongo.
- Wipe cutover: verify talos-dev state; do not run old and new inventory writers together.

## Migration Plan

1. Deploy inventory schema/runtime and retarget reservation CDC using the approved dev wipe.
2. Add ProductCreated contract, products outbox migration/transaction, Kafka topic and connector.
3. Add idempotent inventory creation-event consumption and validate retry behavior.
4. Switch web creation to pending availability; remove EnsureStock RPC/client and compensation-only deletion API. Stop old seeding callers before enabling the new creation flow.
5. Verify publication, consumption, replay, reservation failure, CI configuration, and chart renders. Keep browse dark until #7.
6. Rollback requires a coordinated writer/consumer cutover; do not run both initialization paths. Replay retained creation events when recovering the new flow.

## Open Questions

None for this change. Mongo cutover is #57; Meilisearch, InventoryUpdated projection, and catalog update/delete events are follow-up work in #7.
