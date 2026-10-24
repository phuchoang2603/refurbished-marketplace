## 1. Shared preparation contract

- [x] 1.1 Add `PrepareCheckout` and the shared topic constant in `shared/proto/checkout/v1/` and `shared/messaging/`, including the complete frozen merchant order snapshot and existing correlation context.
- [x] 1.2 Replace sequential initial state values with a fresh PREPARING value; remove and reserve obsolete protobuf fields/state values without reusing wire numbers, and regenerate affected protobuf outputs using the project task.

## 2. Durable Checkout join

- [x] 2.1 Add Checkout-local migrations and queries for immutable preparation identity/version, deadline, branch outcomes/reasons, and irreversible closure; regenerate sqlc outputs.
- [x] 2.2 Update `services/checkout/internal/service/submit.go` to atomically persist PREPARING plus one shared preparation outbox command, retaining buyer-intent idempotency and checkout-ID keys.
- [x] 2.3 Update `results.go` to correlate each branch independently of mutable row version, persist inbox/outcome atomically, and schedule the existing gateway operation once only after both successes.
- [x] 2.4 Update `deadlines.go` and retry queries to close rejected preparation or due preparation claimed by the timeout worker, allowing a successful join until compensation wins the workflow lock, cancel stock, await released/absent, finalize/fence Orders, and await acknowledgment before failure; retry explicit outstanding operations without relying on the latest outbox row.
- [x] 2.5 Update Web and Checkout status mappings and existing test fixtures/helpers for PREPARING; preserve buyer ownership checks and payment/settlement behavior.

## 3. Participant preparation and failure fences

- [x] 3.1 Extend Orders-local checkout intent migrations/queries to persist creation outcomes and absence fences without an orders-row foreign-key dependency; serialize absent-row create/finalize races with a shared transaction-scoped identity lock and regenerate sqlc outputs.
- [x] 3.2 Adapt Orders' command handler to shared preparation, durable business rejection, replayed definitive outcomes, and failed finalization before creation; atomically commit each result outbox with its mutation/fence.
- [x] 3.3 Adapt Inventory's handler to reserve directly from shared preparation while preserving reservation hashes, order-intent locking, all-lines atomicity, and cancel-before-reserve fencing. Acquire contested stock rows in deterministic product-ID order.
- [x] 3.4 Replace initial topic subscriptions in Orders/Inventory startup with the shared preparation topic, retaining distinct stable service groups and existing settlement subscriptions; keep ProductCreated on its separate client/group.

## 4. Production scaling and processing visibility

- [x] 4.1 Declare `checkout.prepare.requested.v1` in Kafka values with one base partition and six production partitions, keeping all other Kafka topics at one partition and retaining existing retention settings; stage retired topic removal after the migration drain.
- [x] 4.2 Set Orders and Inventory to three replicas in the marketplace production overlay; verify group IDs remain identical across replicas of a service and different across services, and inspect database pool/resource budgets for the increased pod count.
- [x] 4.3 Ensure existing logs/traces identify checkout, service, pod, and processing timing; add only missing context necessary to match processing to group assignment evidence.
- [x] 4.4 Render base and production marketplace/Kafka manifests and verify replica counts, topic partitions, subscriptions/config, and existing mesh policy behavior with multiple endpoints.

## 5. Focused behavior verification

- [x] 5.1 Update affected existing command/transition/deadline fixtures and run existing checks for Checkout, Orders, Inventory, shared contracts, and affected Web code using project tooling; retain the existing gateway/webhook and settlement checks.
- [x] 5.2 Exercise isolated fixtures for both result arrival orders, concurrent join completion, duplicate delivery, conflicting correlation, and restart after one branch commits. Verify one effective order, reservation, and gateway payment, with no payment before both successes.
- [x] 5.3 Exercise order rejection after stock success, stock rejection before order creation, preparation recovery timeout, success after the timeout but before the worker closes preparation, cancellation-before-preparation, and late success while compensating. Verify no active hold or pending order remains after acknowledged failure, including concurrent create/finalize on separate replicas.
- [x] 5.4 Exercise multiple checkout IDs competing for a product with ten units, including multi-line contention across partitions; verify no overselling, partial committed holds, or lost retryable work.

## 6. Affected documentation and cutover preparation

- [x] 6.1 Update only affected saga/state/rollout sections in `docs/order-placement.md`, `docs/architecture.md`, and affected service READMEs; document group inspection, multi-key workload evidence, admission pause/drain, and rollback gates.
- [x] 6.2 Update the existing `docs/diagrams/architecture.json` and regenerate its SVG to show one preparation topic feeding two service groups with three production replicas each; keep the editable JSON tracked and preserve the README image/link behavior.
- [x] 6.3 Prepare a concrete edge admission-pause procedure that leaves status reads/webhooks active; identify old outbox/CDC/consumer lag and nonterminal workflow checks required before replacing the protocol. Do not treat an Argo rolling update as a safe cutover by itself.

## 7. Cluster acceptance and retirement

- [x] 7.1 During an authorized rollout, pause new admission, drain old workflows and old command/result delivery, apply migrations/provision the topic, and deploy the coordinated revision before accepting new workflows. Block on unresolved financial/compensation work.
- [ ] 7.2 Verify three healthy production members per service group and six preparation partitions. Submit isolated checkouts spanning all partitions and collect assignments plus pod-correlated processing evidence across all replicas, including overlapping Orders/Inventory work; do not require an exact two-partitions-per-pod split.
- [ ] 7.3 Restart one participant pod under fixture load and verify rebalance, eventual completion, and no duplicate effective mutations; retain evidence separately from local chart/check results.
- [ ] 7.4 Reopen admission after acceptance, remove retired initial topic definitions only after confirmed quiescence, and record the rollback boundary: new-protocol workflows must drain before old binaries can be restored.
