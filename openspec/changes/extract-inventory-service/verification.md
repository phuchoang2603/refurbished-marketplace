# Apply verification

## Local implementation

Product creation commits catalog and ProductCreated outbox records together. The inventory consumer seeds stock transactionally with inbox and seed-intent records. The EnsureStock gRPC and compensation-only DeleteProduct API are removed. Web forwards explicit initial stock and distinguishes pending/unavailable availability from actual zero stock. Checkout still reserves synchronously. Inventory is enrolled in CI lint, path filters, tests, and vulnerability scans.

## Checks

- Affected inventory, products, and web Go suites pass with the Colima `docker` profile. Coverage includes transactional rollback/retry, command-to-Kafka replay, single reservation event, failure release, missing/batch stock reads, pending availability, and independent Kafka consumer groups.
- Existing `TestPaymentService_ExpireDueSessions` passes, verifying expiry produces payment.failed; inventory tests verify that event releases stock idempotently.
- Products tests verify listing/outbox atomicity and retained event identity/payload while no publisher is available. Actual Debezium delivery remains part of live cutover verification.
- Marketplace and Kafka Helm lint/render pass. The rendered products connector targets products_db/public.products_outbox and uses id, aggregate_id, payload, and tracingspancontext. The products.created topic is retained for replay. Inventory CDC targets inventory_db in the rendered configuration.
- OpenSpec strict validation passes. CI configuration includes inventory in each required list/matrix.

## Talos-dev observation — 2026-09-13

Read-only inspection used `/Users/felix/.kube/talos-dev.yaml`.

- Existing web/products deployments run image revision `66605b5a479a77eb9cff255f3b8edb80eafaea71`.
- There is no inventory Deployment.
- There is no products-outbox KafkaConnector or products.created KafkaTopic.
- The live inventory-outbox connector still targets products_db at products-db-rw.ecommerce.svc.

Task 4.4 remains open: the new image/chart revision must be deployed before verifying old seeding is stopped, inventory consumption and creation CDC are active, and no dual initialization writers remain. No deployment or database wipe was performed during this apply session.
