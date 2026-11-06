# Dev benchmark evidence

Talos Proxmox dev; three nodes, 20 vCPU / 64 GiB provisioned. Dataset: 1,000 synthetic listings created through the HTTP application and projected through PostgreSQL outbox → Debezium → Kafka → Meilisearch.

After profiling and adding 20-listing pagination, the storefront passed 100 offered requests/sec (99.7 achieved) at 433.6 ms p95 across 12,001 requests in two minutes, with zero errors or dropped arrivals. The follow-up measurements below distinguish this workload from the original 100-listing pages.

## Initial bootstrap measurements

| Workload   | Duration | Requests | Achieved requests/sec |      p95 |       p99 | Errors | Dropped arrivals | Passed |
| ---------- | -------- | -------: | --------------------: | -------: | --------: | -----: | ---------------: | ------ |
| inventory  | 2m       |   12,001 |                 100.0 |  19.4 ms |   36.3 ms |  0.00% |                0 | True   |
| inventory  | 2m       |   60,000 |                 499.7 |  92.0 ms |  112.1 ms |  0.00% |                0 | True   |
| storefront | 2m       |    3,600 |                  30.0 | 150.6 ms |  201.0 ms |  0.00% |                0 | True   |
| storefront | 2m       |    5,999 |                  49.9 | 826.2 ms | 1181.9 ms |  0.00% |                2 | False  |

Only passing runs with no errors and no dropped arrivals support résumé throughput claims. The 50-RPS storefront attempt is included as a failed threshold result. Warm-ups, earlier stress attempts and rollout traffic are excluded from this table.

## Measurement boundaries

- **storefront:** In-cluster Istio Gateway → web → gRPC services; 45% browse, 35% search, 20% product detail.
- **inventory:** In-cluster InventoryService/GetStock through Istio waypoint → PostgreSQL; checks product ID and available quantity.

## Conditions and limits

- Development load tests, not production traffic or an availability SLA.
- Client-side k6 timings; excludes Cloudflare, WAN, browser rendering, checkout and payment processing.
- 100% application tracing enabled. Inventory has one replica, web has two.
- Tests use 1-CPU Go service limits, a 1-CPU waypoint limit, and 2-CPU / 512-MiB Meilisearch limits.
- PostgreSQL application pools have a maximum of 16 connections per pod.

## Observations

The seeded catalog was verified to contain 1,000 products, inventory rows and search documents. The observed PostgreSQL application pool reached 16 connections under load. Raw run artifacts are excluded from Git; this report retains the results.

The initial 30 requests/sec attempt ran during indexing with the original small resource limits and exhausted Meilisearch memory. The initial inventory smoke test exhausted PostgreSQL connections with unbounded pools. Those runs are excluded from steady-state comparisons. The first 500-RPS inventory attempt dropped 276 arrivals while dynamically allocating clients; the repeat preallocated 200 virtual users and retained the zero-drop threshold.

## Canary validation

The candidate completed metric-gated traffic stages at 5%, 25%, 50% and 100%, with all four analyses successful. The rollout finished Healthy with matching stable and current pod hashes. An earlier analysis rejected insufficient candidate traffic and restored stable traffic to 100%; increasing the metrics warm-up from 45 to 90 seconds allowed a retry without weakening the gates.

## Résumé wording

- Deployed eight Go microservices on Talos Kubernetes with Argo CD, Istio Gateway API and metric-gated Argo Rollouts, validating progressive delivery at 5%, 25%, 50% and 100% traffic and automatic rollback on failed analysis.
- Benchmarked inventory reads at approximately 500 requests/sec over 60,000 requests with 92 ms p95 latency, zero errors and zero dropped arrivals in a three-node development cluster; bounded PostgreSQL connection pools after diagnosing connection exhaustion.

- Profiled and optimized a CQRS storefront with typed search-result decoding and 20-listing pagination; validated 100 requests/sec across 12,001 mixed browse/search/detail requests at 434 ms p95, with zero errors or dropped arrivals in two-minute development tests.

These numbers describe two-minute synthetic dev tests. They do not establish maximum sustained capacity or production availability.

## Storefront profiling follow-up

A repeat with the original 100-listing pages measured 171.9 ms p95 at 30 requests/sec. At 50 requests/sec, preallocating 200 k6 virtual users eliminated dropped arrivals, but p95 was still 742.1 ms. At 75 requests/sec the workload overloaded: p95 reached 9.66 seconds, errors reached 24.4%, and 4,343 arrivals were dropped. These runs used the original decoder and the same dataset and request mix.

VictoriaMetrics showed Meilisearch reaching its 2-CPU limit during the 75-RPS run, with CPU throttling in up to 50.6% of scheduling periods. The node hosting search and Meilisearch averaged about 85.5% non-idle CPU, including 6.6% steal time. Node pressure and VM contention affect these results; the whole cluster's provisioned CPU is not dedicated application capacity.

A 30-second live CPU profile at 30 requests/sec attributed 43.8% of search-service CPU samples to JSON unmarshalling, including 21.0% in the SDK hit conversion path. A separate local benchmark of a 100-hit payload measured 1.55–1.60 ms and 3,318 allocations through the SDK maps, versus 0.61–0.66 ms and 416 allocations decoding directly into typed documents. These are isolated decode timings, not end-to-end speedups.

The search service now decodes raw search responses into typed listing documents. The public catalog also uses 20-listing pagination with a one-hit lookahead instead of fetching and rendering 100 listings per request. Next/previous links retain the query. No cache or tracing reduction was introduced.

CQRS keeps catalog reads off the write database, but read throughput remains bounded by Meilisearch processing, decoding, RPC transfer, HTML rendering and the resources available to those components. Pagination changes the amount of work per request, so its results must be distinguished from the original 100-listing workload.

## Follow-up measurements

Each row is a separate two-minute mixed storefront test on the same 1,000-product dataset. The 50-RPS comparisons preallocate 200 virtual users. The original pages render 100 listings; the paginated pages render 20 and fetch one extra hit to determine whether a next page exists. All runs retain 100% tracing and the existing service resource limits.

| Version            | Cards/page | Offered RPS | Actual RPS | Requests |       p95 |       p99 | Errors | Dropped arrivals | Passed |
| ------------------ | ---------: | ----------: | ---------: | -------: | --------: | --------: | -----: | ---------------: | ------ |
| Original           |        100 |          30 |       30.0 |    3,601 |  171.9 ms |  288.1 ms |  0.00% |                0 | Yes    |
| Original           |        100 |          50 |       50.0 |    6,001 |  742.1 ms | 1013.8 ms |  0.00% |                0 | No     |
| Original           |        100 |          75 |       37.4 |    4,658 | 9655.6 ms | 9869.6 ms | 24.40% |             4343 | No     |
| Typed decoder      |        100 |          50 |       49.9 |    6,001 |  422.3 ms |  692.2 ms |  0.00% |                0 | Yes    |
| Typed + pagination |         20 |          50 |       50.0 |    6,001 |  117.3 ms |  157.2 ms |  0.00% |                0 | Yes    |
| Typed + pagination |         20 |         100 |       99.7 |   12,001 |  433.6 ms |  900.8 ms |  0.00% |                0 | Yes    |
| Typed + pagination |         20 |         150 |      134.1 |   16,309 | 2114.9 ms | 2390.3 ms |  0.00% |             1692 | No     |

Only rows that pass every latency and correctness threshold, with zero errors and zero dropped arrivals, support throughput claims. Mixed-version rollout traffic, sequential diagnostic probes and the first decoder comparison immediately after promotion are excluded. That first comparison had a brief burst of failed responses; the settled repeat is shown above.

These are observations on a shared development cluster, not an isolated causal experiment. Web pod placement changed during the profiling rollout, and host contention varied. Do not attribute the whole latency difference to typed decoding alone, or compare 20-card and 100-card requests as equal work.

Live navigation verification followed all 50 pages through the candidate application, preserved the search query, and recovered all 1,000 seeded product IDs exactly once at 20 cards per page.

The final pagination rollout completed all four metric-gated traffic stages and finished Healthy. After the sweep, benchmark Jobs and script ConfigMaps were removed; final application health and the public HTTP endpoint were verified again.
