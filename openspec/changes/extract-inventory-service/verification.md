# Apply verification

## Local implementation

Product creation commits catalog and ProductCreated outbox records together. The inventory consumer seeds stock transactionally with inbox and seed-intent records. The EnsureStock gRPC and compensation-only DeleteProduct API are removed. Web forwards explicit initial stock and distinguishes pending/unavailable availability from actual zero stock. Checkout still reserves synchronously. Inventory is enrolled in CI lint, path filters, tests, and vulnerability scans.

## Checks

- Affected inventory, products, and web Go suites pass with the Colima `docker` profile. Coverage includes transactional rollback/retry, command-to-Kafka replay, single reservation event, failure release, missing/batch stock reads, pending availability, and independent Kafka consumer groups.
- Existing `TestPaymentService_ExpireDueSessions` passes, verifying expiry produces payment.failed; inventory tests verify that event releases stock idempotently.
- Products tests verify listing/outbox atomicity and retained event identity/payload while no publisher is available. Actual Debezium delivery remains part of live cutover verification.
- Marketplace and Kafka Helm lint/render pass. The rendered products connector targets products_db/public.products_outbox and uses id, aggregate_id, payload, and tracingspancontext. The products.created topic is retained for replay. Inventory CDC targets inventory_db in the rendered configuration.
- OpenSpec strict validation passes. CI configuration includes inventory in each required list/matrix.

## Talos-dev cutover — 2026-09-13 (task 4.4)

Kubeconfigs: `/Users/felix/.kube/talos-dev.yaml` (workloads), `/Users/felix/.kube/talos-gpu.yaml` (Argo CD). Git revision / image tag: `faf5b5cc16ff56402ae2ad88c18e3d2d2c489e15` (`split-runtime`).

### Runtime

- `dev-refurbished-marketplace` sync succeeded after `INVENTORY_APP_PASSWORD` was present in Doppler and ESO synced `inventory-app`.
- Deployments `inventory`, `products`, and `web` run `:faf5b5cc…`. Inventory Deployment Ready; `inventory-db` CNPG healthy; `inventory-migrate` and `products-migrate` Completed.
- `products_db` public tables are only `products`, `products_outbox`, and goose metadata (inventory tables dropped by migration 007). `inventory_db` owns inventory / reservations / inbox / outbox / seed_intents.
- Web env has `INVENTORY_SVC_ADDR=inventory:9097`. CiliumNetworkPolicy `allow-inventory` admits only `app=web` to inventory `:9097` (no products→inventory path). EnsureStock remains an internal inventory operation only; no EnsureStock gRPC / web seeding saga on the cutover images.

### Creation CDC + inventory consumer

- KafkaTopic `products.created` Ready. KafkaConnector `products-outbox` Ready against `products-db-rw` / `products_db` / `public.products_outbox`.
- KafkaConnector `inventory-outbox` Ready against `inventory-db-rw` / `inventory_db` / `public.inventory_outbox` (no longer products_db).
- Inventory pod logs: kafka consumer group `inventory-service` subscribed to `products.created,orders.created,payment.*`.

### Cutover ops notes (dev-only)

- First marketplace sync failed while Doppler lacked `INVENTORY_APP_PASSWORD`; re-sync after ESO created the secret.
- `products-outbox` initially created an empty filtered publication because the connector started before `products_outbox` existed; publication was updated to include `public.products_outbox` and the connector restarted so snapshot/stream could proceed.
- `inventory-outbox` retained stale Connect offsets from the previous products_db inventory CDC path; offsets were reset and the leftover `debezium_inventory_slot` on products_db was dropped so the connector could snapshot/stream on inventory_db.

### End-to-end seed evidence

- `CreateProduct` with `initial_stock=7` → product `39456006-dda8-4816-b00d-eda5844b4a77` + outbox row; after publication fix/snapshot, inventory seeded `available_qty=7`.
- `CreateProduct` with `initial_stock=11` → product `46db0ceb-da56-4b65-9e07-8ef17dcd49c3`; `GetStock` returned `available=11 reserved=0` within seconds via live CDC.
- Inventory inbox contains matching `products.created/<event_id>` rows and seed_intents for both product ids. No second initialization writer: products_db has no inventory tables; only inventory consumes creation events.

Task 4.4 complete: old web stock seeding stopped, creation CDC active, inventory consumer ready, single initialization writer path verified on talos-dev.
