## 1. Meilisearch GitOps

- [ ] 1.1 Add wrapper chart `infra/charts/meilisearch/` (upstream meilisearch-kubernetes, single-node PVC, Doppler ExternalSecret `MEILI_MASTER_KEY`, Cilium CNP 7700 from products + kubelet, no SPIRE)
- [ ] 1.2 Register an app-of-apps child destining `ecommerce` at sync wave `2`; do not template an `ecommerce` Namespace
- [ ] 1.3 Document Doppler key, GitOps row, and Cilium path in secrets/gitops/cilium docs

## 2. Products projector and SearchProducts

- [ ] 2.1 Add `shared/testutil/meilisearch` Testcontainers helper; enroll it in `go.work`, lint globs, and products path-filter fan-out
- [ ] 2.2 Replace `ListProducts` with `SearchProducts` in proto (text, optional merchant filter, offset/limit); regenerate clients
- [ ] 2.3 Wire products to Meilisearch and Kafka (`MEILI_URL` + master key, `KAFKA_BOOTSTRAP_SERVERS`); wait-for-meili `/health`; fail SearchProducts if Meili is down (no Mongo list fallback)
- [ ] 2.4 Consume `products.created` in group `products-search-product-created`; upsert catalog fields (no `initial_qty`); ensure index settings; rebuild from Mongo listings
- [ ] 2.5 Cover SearchProducts browse/merchant filter and create→index against testutil Meili (and Kafka as needed)

## 3. Web

- [ ] 3.1 Public catalog calls SearchProducts with empty query; render hits without stock; use the products unavailable page on search failure
- [ ] 3.2 Seller list calls SearchProducts with merchant filter; remove ListProducts callers and in-process merchant filtering
- [ ] 3.3 Leave PDP on GetProductByID + inventory GetStock

## 4. Cutover

- [ ] 4.1 Update `docs/catalog-inventory-search.md` for the Meili projection, independent consumer group, and browse vs PDP
- [ ] 4.2 Set Doppler `MEILI_MASTER_KEY`; deploy Meili then products/web; rebuild index; verify create → browse and seller list on talos-dev
