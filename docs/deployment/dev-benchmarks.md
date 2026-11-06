# Dev benchmark evidence

Talos Proxmox dev; three nodes, 20 vCPU / 64 GiB provisioned. Dataset: 1,000 synthetic listings created through the HTTP application and projected through PostgreSQL outbox → Debezium → Kafka → Meilisearch.

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

- Validated browse, search and product-detail pages against 1,000 seeded listings at 30 requests/sec across 3,600 requests, with 151 ms p95 latency and zero errors or dropped arrivals.

These numbers describe two-minute synthetic dev tests. They do not establish maximum sustained capacity or production availability.
