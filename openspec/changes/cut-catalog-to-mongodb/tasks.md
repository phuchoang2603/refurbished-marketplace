## 1. Products Mongo persistence

- [x] 1.1 Add `shared/testutil/mongo` replica-set Testcontainers helper; enroll it in `go.work`, lint globs, and products path-filter fan-out
- [x] 1.2 Add official mongo-go-driver to `services/products`, load replica-set URI/user/password (host + `mongodb-catalog-app`), and fail startup without Mongo
- [x] 1.3 Implement `listings` and `catalog_outbox` access with `Session.WithTransaction` for CreateProduct (EventRouter field names, `initial_qty` and tracing on the outbox payload)
- [x] 1.4 Point GetProductByID, GetProductsByIDs, and ListProducts at Mongo; remove sqlc/goose/Postgres from the products module
- [x] 1.5 Cover create atomicity/rollback and retained event identity against `shared/testutil/mongo` (replica set, not standalone)

## 2. Debezium Mongo publication

- [x] 2.1 Add Debezium MongoDB connector (same 3.5.0.Final line as postgres) to `infra/docker/connect-debezium.Dockerfile`
- [x] 2.2 Keep Kafka topic `products.created`; replace the Postgres products-outbox connector with a Mongo outbox connector (EventRouter identity/key/payload/tracing unchanged)
- [x] 2.3 Allow Kafka Connect to read `mongodb-catalog-app` in `ecommerce` (existing secret Role pattern)

## 3. GitOps, mesh, and secrets

- [x] 3.1 Mount `mongodb-catalog-app` on the products Deployment; drop products CNPG, migrator, and `products-app` PG ExternalSecret from the marketplace chart
- [x] 3.2 Allow Kafka Connect (and products, kubelet) to Mongo 27017 without SPIRE; keep default-deny for other identities
- [x] 3.3 Update secrets/GitOps/Cilium docs for products Mongo auth, Connect’s Mongo path, and the retired products Postgres path

## 4. Docs and cutover

- [x] 4.1 Update `docs/catalog-inventory-search.md` for Mongo listing+outbox and Debezium Mongo CDC
- [x] 4.2 Wipe talos-dev SQL listings; deploy Mongo-backed products and the Mongo connector; verify create → CDC → inventory seed and seller list; then delete leftover `products-db` if still present
