## Context

See `proposal.md` for motivation. Submission already allocates checkout/order IDs and freezes the merchant-scoped item snapshot. Today `submit.go` writes a create-order command; `results.go` requests reservation only after OrderCreated. The current result guard compares operation identity and workflow version against one expected operation for the current state. That guard would incorrectly reject the second parallel result after the first changes the row version.

Debezium keys outbox records by `aggregate_id`, currently checkout ID. The shared Kafka consumer already processes fetched partitions concurrently and records within each partition sequentially. Each service starts one command consumer client per process; multiple clients in the same service group can distribute partitions without adding an in-process worker pool.

Inventory already serializes order reservation intent and uses PostgreSQL stock row locks, with cancellation-before-reserve tombstones. Orders currently requires an existing checkout-order binding for finalization and therefore cannot fence finalization-before-create. Existing Orders specs also require item data on an order-created result that is currently an empty protobuf payload with correlation context; the shared command becomes the authoritative preparation snapshot and the delta corrects that mismatch.

## Goals / Non-Goals

**Goals:** make order creation and reservation independently executable, retain a durable payment gate and compensating failure protocol, and demonstrate partition distribution across three production pods per participant.

**Non-Goals:** retain the scope exclusions in the proposal. Do not add per-item consumers, split one merchant order across transactions, or depend on cross-topic delivery order. Do not promise lower end-to-end latency or broker high availability.

## Decisions

### One command, two independently delivered branches

Add a `PrepareCheckout` protobuf payload on `checkout.prepare.requested.v1` containing buyer ID, merchant ID, total, currency, and the full frozen line snapshot. Keep checkout/order IDs in MessageContext. Submission commits the workflow in a new PREPARING state and one outbox row together. Orders and Inventory subscribe in their existing, distinct command groups, replacing their respective initial topic subscriptions. Preserve the current finalization and stock settlement topics and result topics.

Orders creates its pending order; Inventory reserves all lines directly from the same command. Neither calls or awaits the other. Checkout continues to receive the two typed result streams. Two separate commands could also permit parallel execution, but the shared command is the requested topology and avoids duplicate initial publication. Participant-specific payload needs now share a contract, so both handlers must validate the relevant fields without mutating the snapshot.

Remove the old CreateOrder/ReserveStock message alternatives and topic constants after cutover; reserve removed protobuf field numbers/names and never reuse them. Add a fresh enum value for PREPARING and reserve removed sequential state values/names. Update generated code and Web's buyer-progress mapping together. The browser continues to receive pending progress until the durable join completes.

### Durable join with stable preparation correlation

Checkout owns service-local columns for preparation operation identity, immutable preparation workflow version, deadline, closure decision, and independent Orders/Inventory outcomes (pending, succeeded, rejected), with rejection reasons. There is one preparation attempt per accepted intent; retries retain that attempt's identity and frozen payload. The preparation operation can remain deterministically derived from checkout ID. Reuse MessageContext.workflow_version as the immutable preparation version for preparation commands/results, storing it separately from the mutable checkout row version.

For each result: lock the checkout, verify checkout/order binding, operation identity, preparation version and expected participant/result type, deduplicate in the inbox, and persist the branch outcome in the same transaction. A valid sibling result is accepted regardless of intervening row-version increments. Identical repeated outcomes have no extra effect; conflicting definitive outcomes preserve prior facts and produce a reconciliation exception.

Only the transaction that observes both successes while preparation remains open transitions to CREATING_PAYMENT and inserts the existing durable gateway operation. Retain gateway order-ID idempotency and uniqueness constraints. Mutable row version remains a concurrency guard; post-preparation commands use their own recorded operation/version. No database transaction spans a network call. Inferring branch completion from the aggregate state or from topic ordering would lose valid sibling results and is rejected.

### Close preparation before compensation

Use a persisted two-minute recovery timeout, matching the existing preparation interval. Passing the timeout makes a still-preparing workflow eligible for the deadline worker; it does not itself close preparation or invalidate a successful result. The deadline worker and result handlers serialize their decisions under the same checkout row lock. If both successes commit before timeout compensation closes preparation, including after the timeout becomes due, payment preparation proceeds under its separate existing deadline. If timeout compensation or a definitive rejection closes preparation first, closure is irreversible and no late success can open payment creation. Gateway creation deadlines and the gateway payment expiry/reconciliation policy remain separate and unchanged.

On closed preparation, send the existing stock cancellation even if Inventory's outcome is unknown. Await released/absent acknowledgment, then send failed order finalization even if Orders' creation outcome is unknown. Await Orders' failed/fenced-absent acknowledgment before terminal FAILED. This keeps the existing stock-before-order settlement rule while covering both partial-success directions. Late results are recorded as needed for reconciliation and never reopen preparation. Unacknowledged compensation stays durable and retryable rather than being declared complete by time alone.

Retry work must identify the outstanding operation explicitly. `retryLastCommand` currently depends on the latest outbox row and row version; replace that assumption for preparation and compensation with persisted operation/payload identity. Retransmissions can use fresh delivery message IDs but retain operation, attempt version, and payload. Participant intent deduplication must return the existing definitive outcome on such redelivery. Re-publishing a shared command can reach an already successful branch; that branch must not mutate twice.

### Orders owns the missing creation fence

Extend Orders' local checkout intent persistence to represent pending creation, rejected creation, and failed absent fences without requiring an existing orders row. Bind every intent to checkout ID and order ID; store the deterministic preparation payload hash when known. Serialize create and finalize by the same transaction-scoped identity lock, including when no intent row exists, and retain uniqueness constraints. An existing row lock alone cannot protect absent-row races.

Failed finalization after stock release either marks an existing order failed or writes a durable absence fence. Commit the finalization result in the same transaction. A delayed preparation finding that fence cannot create a pending order and reports a definitive rejected preparation outcome. Rejected preparation is also persisted with its outbox result. Paid finalization still requires an existing order and rejects contradictory terminal outcomes. Persistence belongs under `services/orders/db/`; do not introduce a shared cross-service transaction.

### Inventory keeps its stock correctness boundary

Adapt the shared preparation snapshot into Inventory's existing reservation intent/hash and reservation transaction. Keep order-level intent locking, all-lines atomicity, inbox/outbox writes, and released/absent tombstones. A late preparation finding a cancellation fence returns rejection and never holds stock. Stock row locks, rather than checkout-key partition ordering, prevent overselling when different checkouts contend for the same product. Use a deterministic product-ID lock acquisition order for multi-line reservations to reduce deadlocks; transient database failures remain retryable.

Keep ProductCreated on its independent consumer client/group. Three Inventory replicas also scale that consumer, but its separate topic partition count controls useful parallelism; this change makes no three-way seed-consumption guarantee.

### Production topology and evidence

Set `services.orders.replicas` and `services.inventory.replicas` to 3 in `infra/charts/refurbished-marketplace/values-prod.yaml`. Declare the shared preparation topic with one partition in Kafka base values and six with the existing seven-day retention in production. Keep checkout-ID keys and stable service group IDs across pods; pod-specific group IDs would incorrectly broadcast within a service. Only the shared preparation topic has six production partitions; all other declared Kafka topics have one. This concentrates three-member parallelism on preparation and reduces partition overhead for results and settlement commands. Single-partition topics have one assigned consumer per group, while the other replicas still process preparation. Local/base settings remain small by explicit planning assumption. Existing Kafka topics cannot decrease their partition count in place; apply the reduced counts by recreating them during the agreed fresh start, preserving no required messages or workflows. This configuration edit does not itself reset live topics.

Use group descriptions for member/partition assignments and existing correlated logs/traces to identify service, pod, checkout, and handler duration. Add only missing processing context needed to demonstrate overlap. Use multiple checkout IDs spanning all partitions; a single key cannot demonstrate partition-level parallelism. Check that all three pods per service process assigned work and that Orders/Inventory spans can overlap. Consumers use the range balancer, which splits each topic evenly across the three members; the default cooperative-sticky balancer balances total partitions across topics and can leave a replica without any preparation partitions. Extend operational instructions only in affected saga/deployment sections, and update the tracked architecture JSON plus its SVG without exporting PNG or introducing Excalidraw.

## Risks / Trade-offs

- Shared payload couples initial participant contracts → version it explicitly and keep service-local interpretation and validation.
- Inventory can hold stock before an order exists → a durable join, recovery timeout worker, cancellation fence, and Orders absence fence cover partial success.
- Three replicas increase database connections and contention → inspect pool/resource settings, retain transactional locks, and verify hot-product behavior without assuming linear speedup.
- Cross-topic messages and rebalances can reorder or redeliver work → correlate operations independently of the row revision and serialize participant decisions in PostgreSQL.
- Kafka remains single-broker RF1 in production → this change demonstrates consumer parallelism, with no new broker availability claim.
- Coordinated internal protocol replacement requires a checkout admission pause → drain existing workflows and delivery backlogs before retiring the old subscriptions.

## Migration Plan

1. Render both chart configurations, regenerate protobuf/sqlc outputs, run affected existing checks, and verify the failure/reordering matrix against isolated fixtures before deployment.
2. Pause new checkout admission at the edge using an operational gate while keeping status reads, webhooks, consumers, and workers running. Drain accepted workflows through stock/order settlement; unresolved financial or compensating workflows block cutover. Verify old command/result lag is drained and Debezium has published the relevant outboxes, not merely that buyer requests stopped.
3. Apply additive service-local schema migrations and provision the new shared topic. Remove the retired initial-command topics from the Kafka chart in the same change.
4. Stop the old Checkout producer/workers and deploy the coordinated Checkout/Orders/Inventory/Web revision with admission still paused. Start and verify both participant groups on the shared topic, three production members per group, and compatible result consumers before admitting new work. Argo rollout order alone is insufficient as a drain mechanism.
5. Run isolated fixture acceptance and compensation checks, reopen checkout admission, and observe assignments and multi-key overlap. Old protobuf/state handlers are removed rather than retained as compatibility paths.
