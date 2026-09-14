## 1. Meilisearch GitOps

- [x] 1.1 Add wrapper chart `infra/charts/meilisearch/` (upstream meilisearch-kubernetes, single-node PVC, Doppler ExternalSecret `MEILI_MASTER_KEY`, Cilium CNP 7700 from search + kubelet, no SPIRE)
- [x] 1.2 Register an app-of-apps child destining `ecommerce` at sync wave `2`; do not template an `ecommerce` Namespace
- [x] 1.3 Document Doppler key, GitOps row, and Cilium path in secrets/gitops/cilium docs

## 2. Search service

- [ ] 2.1 Add `services/search` module, GHCR image, Helm Deployment/Service, go.work/lint/CI/govulncheck/release-images enrollment
- [ ] 2.2 Add `shared/testutil/meilisearch` Testcontainers helper; enroll it in `go.work`, lint globs, and search path-filter fan-out
- [ ] 2.3 Add search proto `SearchProducts` (text, optional merchant filter, offset/limit); remove products `ListProducts`; regenerate clients
- [ ] 2.4 Wire search to Meilisearch and Kafka; wait-for-meili `/health`; consume `products.created` as `search-product-created`; upsert catalog fields (no `initial_qty`); no Mongo client or rebuild
- [ ] 2.5 Cover SearchProducts browse/merchant filter and create→index against testutil Meili and Kafka

## 3. Web and products

- [ ] 3.1 Public catalog calls search SearchProducts with empty query; render hits without stock; catalog unavailable page on search failure
- [ ] 3.2 Seller list calls search SearchProducts with merchant filter; remove ListProducts callers and in-process merchant filtering
- [ ] 3.3 Leave PDP on products GetProductByID + inventory GetStock; products has no Meili or Kafka consumer

## 4. Cutover

- [ ] 4.1 Update `docs/catalog-inventory-search.md` for the search read-side service, independent consumer group, and browse vs PDP
- [ ] 4.2 Set Doppler `MEILI_MASTER_KEY`; wipe talos-dev catalog/search data; deploy Meili then search/products/web; verify create → browse and seller list
