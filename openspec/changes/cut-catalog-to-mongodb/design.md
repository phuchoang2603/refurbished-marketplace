## Context

See proposal.md for motivation. Products today commits SQL `products` + `products_outbox` in one Postgres transaction; Debezium Postgres + EventRouter publishes `products.created`. Inventory already consumes that topic. Mongo replica set and CNP (`products` → 27017) already exist; products only has unused `MONGO_ADDR`. Official Go driver v2 runs multi-document transactions with `Session.WithTransaction`; operations inside the callback MUST use the session context (`mongo-go-driver`). A 1-member replica set is enough for those transactions and for Mongo change streams.

## Goals / Non-Goals

**Goals:**

- Mongo as catalog store and outbox store; CreateProduct returns after the replica-set transaction commits.
- Debezium Mongo connector + EventRouter publishes retained `ProductCreated` protobufs to the existing topic (product-id key, tracing metadata).
- `shared/testutil/mongo` replica-set helper, same shape as postgres/kafka/redis testutils.
- Remove products Postgres (cluster, migrator, sqlc, PG connector).
- Seller `ListProducts` from Mongo; public browse stays dark in web.

**Non-Goals:**

- Changing inventory or web gRPC call graph.
- Meilisearch, update/delete catalog events, `InventoryUpdated`.
- In-process Kafka publisher in products (do not run alongside CDC).
- A second Mongo user if the existing `catalog` SCRAM secret can be mounted on products and Connect.

## Decisions

### 1. Catalog and outbox collections in one Mongo transaction

Database `catalog` (already on the Community CR). Collections: `listings` (`_id` = product UUID string, v1 fields, `created_at`/`updated_at`) and `catalog_outbox` (`_id` = event UUID, `aggregate_id`, `event_type`, protobuf `payload`, `tracingspancontext`). Field names MUST match the existing EventRouter mapping (`id`, `aggregate_id`, `payload`, `event_type`, `tracingspancontext`). CreateProduct inserts both in `Session.WithTransaction`. `initial_qty` lives only on the outbox payload, not as live listing stock. Do not mark-published in the application; CDC watches inserts.

**Rationale:** Same atomicity as today’s PG TX. Driver retries transient commit errors on the session.

**Alternatives considered:** Mongo listing + SQL outbox (crash window). Change-stream worker in products (second publisher path). In-process relay after commit (new producer, duplicates house CDC).

### 2. Debezium Mongo connector, not a products producer

Keep Kafka topic `products.created`. Replace the Postgres products connector with `io.debezium.connector.mongodb.MongoDbConnector` on collection `catalog.catalog_outbox`, EventRouter SMT and converters unchanged from other outbox connectors. Add `debezium-connector-mongodb` (same 3.5.0.Final line as postgres) to `connect-debezium`. Kafka chart templates MUST select connector class/config per entity (Postgres entities stay as they are). Connect reads `mongodb-catalog-app` via the existing ecommerce secret Role.

**Rationale:** Publication, identity retention, and tracing stay on Connect like orders/payment/inventory. Products stays a catalog service.

**Alternatives considered:** In-process relay (avoids a Connect plugin, but splits publication styles and needs a producer). Dual relay+CDC (double `ProductCreated`).

### 3. Reuse the existing catalog SCRAM secret

Mount `mongodb-catalog-app` on the products Deployment (`catalog-mongodb-svc:27017`, auth DB `catalog`). Point the Mongo connector at the same secret. Remove `products.db`, products migrator, and `products-app` PG ExternalSecret.

**Rationale:** P0 already provisioned that user for this replica set.

**Alternatives considered:** A dedicated products-mongo Doppler key (more secrets, no extra isolation on a 1-member RS).

### 4. Mongo CNP allows Connect

Extend `infra/charts/mongodb` ingress to Kafka Connect pods (kafka namespace / documented Strimzi labels) on 27017 without SPIRE, alongside `app: products` and `fromEntities: host`. Unknown callers stay denied when enforce is on.

**Rationale:** Default-deny would drop CDC even if the connector is Healthy.

### 5. Shared Mongo Testcontainers module

Add `shared/testutil/mongo` using `testcontainers-go/modules/mongodb` with replica-set mode (standalone cannot run the create transaction). Helper returns a driver-ready URI/client; products tests replace `shared/testutil/postgres`. Enroll the module in `go.work`, lint globs, and path-filter fan-out → products. Do not run Connect in unit tests; assert listing+outbox atomicity in Mongo. Live CDC remains talos-dev verification, same as today’s PG Debezium tests.

**Rationale:** Matches postgres/kafka/redis testutil layout; replica set is required, not optional.

### 6. Drop PG publication in the same GitOps cutover

Remove the Postgres products connector. Disable marketplace products CNPG. Wipe shop-dev listings; no backfill. Do not run SQL CreateProduct and Mongo CreateProduct together.

**Rationale:** Dual writers would double-publish `ProductCreated`.

**Alternatives considered:** Copy SQL rows into Mongo first (not worth it after the extract wipe). Leave empty `products_db` (orphan cluster).

### 7. Docs

Update `docs/catalog-inventory-search.md` for Mongo listing+outbox and Mongo Debezium; keep OpenSpec deltas as the requirement source. Document Connect’s Mongo secret and CNP in GitOps/secrets/Cilium notes.

## Risks / Trade-offs

- [CDC lag] → Same as today: PDP pending stock, checkout ReserveStock fails closed until inventory seeds.
- [Connect cannot reach Mongo] → CNP from Connect; verify with a live create after cutover.
- [Mongo EventRouter field mapping] → Keep outbox document keys identical to the SQL outbox columns the SMT already uses.
- [Standalone Mongo in tests] → Fail closed; testutil must start a replica set.
- [Dropping products_db] → Irreversible without restore; cut over only after Mongo create/read and live `products.created` seed are verified.
- [Shared SCRAM user] → Products, Connect, and operator tooling share `catalog`; acceptable on talos-dev.
- [Connect image size] → Second Debezium plugin next to postgres; pin the same version.

## Migration Plan

1. Add Mongo driver, collections, transactional create/reads, `shared/testutil/mongo`, Helm env/secret/CNP, Connect Mongo plugin + connector; ship products without a PG writer.
2. Deploy; verify CreateProduct → Mongo outbox → Kafka → inventory seed on talos-dev.
3. Remove leftover PG connector/cluster/migrator/sqlc/goose if any remain.
4. Rollback is a previous image plus restored `products_db`/PG connector only if the cluster was not yet deleted; after drop, replay is Mongo outbox → CDC.

## Open Questions

None. Publisher is Debezium Mongo; tests use `shared/testutil/mongo`; Mongo user is the existing `catalog` SCRAM secret.
