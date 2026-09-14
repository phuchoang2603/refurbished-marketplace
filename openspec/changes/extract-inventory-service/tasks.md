## 1. Proto and module

- [x] 1.1 Finalize inventory.v1 with GetStock, GetStocksByIDs, and ReserveStock; remove EnsureStock RPC and regenerate
- [x] 1.2 Keep ReserveStock removed from products.v1; add explicit optional initial quantity to CreateProduct; remove compensation-only DeleteProduct RPC/client/service/query and regenerate
- [x] 1.3 Add services/inventory Go module and workspace tidy
- [x] 1.4 Define ProductCreated on products.created with stable event id, schema version, occurrence time, product id/version, catalog fields, and explicit initial quantity; generate protobufs and register messaging constants

## 2. Inventory runtime

- [x] 2.1 Goose schema: inventory, reservations, inbox, outbox (no FK to products); sqlc
- [x] 2.2 Implement internal EnsureStock, stock reads, ReserveStock with existing reservation semantics
- [x] 2.3 Move reservation Kafka consumer + outbox publish to inventory (:9097 gRPC)
- [x] 2.4 Helm: inventory Deployment/Service, CNPG Cluster, ESO, Cilium allow web only, Doppler key docs
- [x] 2.5 Consume ProductCreated in inventory's independent group; validate explicit non-negative initial quantity; transactionally record event identity and seed stock before acknowledgement
- [x] 2.6 Preserve seed intent for product-level replay/conflict detection; duplicates must never reset stock after reservations; roll back transient failures and retry through the existing consumer error path

## 3. Products and web cutover

- [x] 3.1 Keep products catalog-only reads and no inventory client; add products outbox migration/sqlc and atomic listing + ProductCreated persistence; validate missing/negative quantity while accepting explicit zero
- [x] 3.2 Replace web EnsureStock/compensating delete with CreateProduct carrying initial quantity and a created/processing response; remove obsolete saga code and fakes
- [x] 3.3 Web checkout ReserveStock → inventory; add-to-cart stamps from form (no GetProduct)
- [x] 3.4 Retarget Debezium/Kafka chart connectors from products_db inventory_outbox to inventory_db
- [x] 3.5 Provision products.created topic and products outbox CDC connector with secret/RBAC access, product-id key, stable event id, and tracing metadata; wire inventory subscription without changing reservation topics
- [x] 3.6 PDP distinguishes pending missing stock, unavailable read failures, and actual zero stock; disable purchase controls for pending/unavailable availability; keep browse dark until #7

## 4. Docs and migration

- [x] 4.1 Update proposal/design/specs and docs/catalog-inventory-search.md diagrams for ProductCreated, independent consumers, durable publication, and future projection ordering
- [x] 4.2 Move baseline reserve idempotency and Kafka tests to inventory (additional coverage below remains open)
- [x] 4.3 Wipe talos-dev inventory/catalog stock data; shop list may stay empty
- [x] 4.4 Verify coordinated cutover: old web stock seeding stopped, creation CDC active, inventory consumer ready, no dual initialization writers; record dev verification evidence

## 5. CI enrollment

- [x] 5.1 Add inventory to .github/workflows/ci.yml GO_MODULE_GLOBS for lint
- [x] 5.2 Add inventory path-filter output and filters for services/inventory, shared dependencies, go.work, and CI workflow changes
- [x] 5.3 Add inventory to the service test matrix and vulnerability-scan matrix, including scheduled scans

## 6. Inventory scenario coverage

- [x] 6.1 Test successful command ReserveStock followed by orders.created replay; assert no second hold and exactly one inventory.reserved outbox record, including repeated deliveries
- [x] 6.2 Test payment failure and expiry release paths, including replay; assert stock is restored once and reservation status is released
- [x] 6.3 Test command full-order failure with a missing/insufficient line; assert no partial hold/reservation and the required inventory.reservation_failed signal
- [x] 6.4 Test stock reads for existing and missing ids, batch omission of missing ids, and batch limits
- [x] 6.5 Fix TestKafkaOrdersCreatedFailure_EndToEnd to wait for a committed inbox/outbox record proving consumption before asserting unchanged stock; assert the failure event
- [x] 6.6 Test ProductCreated first seed (including explicit zero), absent/negative quantity rejection, duplicate event replay after stock changes, conflicting seed intent, rollback on processing failure, and retry after commit
- [x] 6.7 Test command ReserveStock failure, later stock seed, and `orders.created` replay; assert no hold, no `inventory.reserved`, and command retry remains failed

## 7. Pivot verification

- [x] 7.1 Test atomic products listing/outbox persistence and rollback; verify delayed CDC publication retains original event identity and product key
- [x] 7.2 Test web creation without inventory calls/compensation, explicit quantity validation, pending/unavailable PDP behavior, and checkout before initialization preventing payment
- [x] 7.3 Run affected Go suites with Docker available, render/lint marketplace and Kafka charts, validate OpenSpec, and verify CI inventory enrollment

Meilisearch implementation, InventoryUpdated publication/projection, and catalog update/delete events remain #7; Mongo catalog cutover remains #57.
