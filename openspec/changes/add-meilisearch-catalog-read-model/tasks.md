## 1. Phase 0 — Meilisearch via app-of-apps

- [ ] 1.1 Add `infra/charts/meilisearch/` Helm wrapper (upstream Meilisearch chart, persistence, resource limits)
- [ ] 1.2 Wire `MEILI_MASTER_KEY` via External Secrets / Doppler (no plaintext key in Git)
- [ ] 1.3 Register Meilisearch child Application in `infra/argocd/app-of-apps` for local + staging values
- [ ] 1.4 Ensure network path from `ecommerce`/products to Meilisearch Service DNS
- [ ] 1.5 Set products env for Meilisearch URL + API key (marketplace chart or secret wiring)
- [ ] 1.6 Document deploy topology (app-of-apps platform chart; not sidecar) and optional port-forward for search DX
- [ ] 1.7 Verify Argo Healthy and Meilisearch health endpoint reachable from the products network path

## 2. Phase 1 — Catalog outbox and projection

- [ ] 2.1 Add `products_outbox` goose migration (+ tracing column pattern if matching inventory outbox)
- [ ] 2.2 Add SQLC queries for catalog outbox insert/list helpers as needed
- [ ] 2.3 Emit catalog outbox row inside `CreateProduct` transaction
- [ ] 2.4 Add Debezium connector / table include for `products_outbox` in kafka chart values
- [ ] 2.5 Version Meilisearch products index settings in-repo (searchable + filterable attributes)
- [ ] 2.6 Implement projector consumer: catalog events → Meilisearch upsert/delete
- [ ] 2.7 Wire products service config for Meilisearch URL + API key
- [ ] 2.8 Add documented reindex/rebuild from PostgreSQL snapshot (script or one-shot command)
- [ ] 2.9 Document eventual consistency lag and failure/retry behavior
- [ ] 2.10 Decide and implement or explicitly defer reservation → `available_qty` projection (document outcome)

## 3. Phase 2 — Search API and web

- [ ] 3.1 Add/extend products protobuf for `SearchProducts` (query, filters, pagination)
- [ ] 3.2 Implement products gRPC handler querying Meilisearch
- [ ] 3.3 Route storefront list/browse through Meilisearch-backed API (empty query = browse)
- [ ] 3.4 Update web public catalog handler to call search/list API (including query param)
- [ ] 3.5 Update web seller product list to pass merchant filter server-side
- [ ] 3.6 Generate proto stubs (`generate-proto`) and sync workspace modules if needed (`tidy`)

## 4. Verification

- [ ] 4.1 Phase 0: Meilisearch Argo app Healthy; products can reach Service DNS
- [ ] 4.2 Phase 1: create product → document appears in Meilisearch; run rebuild and spot-check parity
- [ ] 4.3 Phase 2: search by name substring; filter by merchant; verify web catalog + seller list
