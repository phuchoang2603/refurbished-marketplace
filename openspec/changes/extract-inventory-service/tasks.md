## 1. Proto and module

- [ ] 1.1 Add `shared/proto/inventory/v1` (EnsureStock, GetStock, GetStocksByIDs, ReserveStock) and generate
- [ ] 1.2 Remove ReserveStock from `products.v1`; add listing delete if missing; generate; fix compile breaks
- [ ] 1.3 Add `services/inventory` Go module and workspace tidy

## 2. Inventory runtime

- [ ] 2.1 Goose schema: inventory, reservations, inbox, outbox (no FK to products); sqlc
- [ ] 2.2 Implement EnsureStock, stock reads, ReserveStock with existing reservation semantics
- [ ] 2.3 Move reservation Kafka consumer + outbox publish to inventory (`:9097` gRPC)
- [ ] 2.4 Helm: inventory Deployment/Service, CNPG Cluster, ESO, Cilium allow **web only**, Doppler key docs

## 3. Products and web cutover

- [ ] 3.1 Products: catalog-only Create/Get/batch; no inventory client; compensating Delete listing RPC
- [ ] 3.2 Web create: products then EnsureStock; delete listing on failure
- [ ] 3.3 Web checkout ReserveStock → inventory; add-to-cart stamps from form (no GetProduct)
- [ ] 3.4 Retarget Debezium/kafka chart connectors from products_db inventory_outbox to inventory_db

## 4. Docs and verify

- [x] 4.1 Architecture doc `docs/catalog-inventory-search.md` with mermaid (products, inventory, Meili)
- [ ] 4.2 Inventory tests: reserve idempotency and Kafka path (existing cases moved)
- [ ] 4.3 Wipe talos-dev inventory/catalog stock data; shop list may stay empty
