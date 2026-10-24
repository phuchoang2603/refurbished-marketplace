## Why

Checkout currently waits for Orders before requesting Inventory even though the accepted snapshot already contains everything both services need. A shared Kafka preparation command lets these branches run independently and makes consumer-group parallelism observable across three replicas per service.

## What Changes

- Publish one durable `checkout.prepare.requested.v1` command per accepted checkout, consumed by Orders and Inventory in distinct service groups.
- Keep Checkout as the durable saga orchestrator: record both preparation outcomes and create payment only after both succeed.
- Use a two-minute recovery timeout, allowing a successful join until compensation wins the workflow lock. Fence late preparation after rejection or timeout compensation starts, including failed order finalization before order creation; retain acknowledged stock compensation before failed order finalization.
- **BREAKING**: replace the separate initial create-order and reserve-stock commands and sequential preparation states. Coordinate a drained cutover rather than supporting mixed protocols.
- Configure three Orders replicas and three Inventory replicas in the production Helm overlay, with six preparation-topic partitions and one partition for every other Kafka topic. Keep base/development defaults at one replica/partition.
- Verify assignments, overlapping processing, replay safety, and compensation using existing traces and Kafka tooling.

Non-goals: choreography, removing Checkout's database, adding checkout mutation RPCs, changing payment integration, item-keyed Kafka partitioning, increasing Kafka broker replication, or introducing a stock projection.

## Capabilities

### New Capabilities

- `checkout-consumer-scaling`: production partition and service-group configuration, with observable distribution across three replicas.

### Modified Capabilities

- `checkout`: atomic shared preparation publication, durable parallel-result join, and deadline compensation.
- `orders`: shared-command consumption, correlated results, and durable finalization-before-create fencing.
- `inventory`: shared-command reservation independent of Orders, retaining stock locking, idempotency, and cancellation fencing.

## Impact

Affected areas: Checkout protobufs and generated clients, `shared/messaging`, service-local migrations/queries and command/result/deadline handlers, consumer startup, Kafka topic values, marketplace production replica values, and the existing architecture diagram's affected flow labels. No new external dependency is required. Buyer submission remains asynchronous; public progress mapping must account for the new preparing state. Existing specs contain sequential-reservation wording and an order-created payload requirement inconsistent with the current empty correlated result; the deltas reconcile those affected requirements explicitly.
