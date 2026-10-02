## 1. Contracts and ownership

- [x] 1.1 Define versioned Checkout command/result protobuf contracts and topic names in `shared/proto/checkout/v1/` and `shared/messaging/`; verify proto codegen succeeds and generated Go types include correlation, operation, and checkout identifiers.
- [x] 1.2 Define the Checkout gRPC submit/status API and generate clients; verify an authenticated status request cannot read another buyer's workflow in a focused service test.
- [x] 1.3 Document the state/transition and command ownership matrix next to the Checkout module; verify every result and deadline in `specs/checkout/spec.md` maps to an allowed transition or reconciliation state.

## 2. Checkout service and persistence

- [x] 2.1 Scaffold `services/checkout/` as a Go workspace module with configuration, gRPC server, and Kafka runner; verify `go test ./...` succeeds in the module.
- [x] 2.2 Add service-local goose migrations, sqlc queries, and indexes for immutable checkout snapshots, unique buyer intent keys, saga version/deadline, inbox, outbox, and exception records; verify migrations and sqlc codegen on a disposable Postgres database.
- [x] 2.3 Implement idempotent SubmitCheckout and buyer-authorized GetCheckout; verify same-key identical payload returns one checkout id and conflicting payload/other buyer is rejected.
- [x] 2.4 Implement the core order → stock → session → payment-ready transitions with transactional inbox, state update, and outbox command; verify redelivery, stale result, and out-of-order result scenarios in focused tests.
- [x] 2.5 Implement payment success → stock commit → paid finalization and definitive payment failure → stock release → failed finalization; verify order remains pending until the corresponding stock settlement acknowledgement in focused tests.
- [x] 2.6 Implement a persistent deadline worker with row locking, retries, and restart recovery for order, stock, session, and payment stages; verify expired work is picked once effectively after worker restart.
- [x] 2.7 Implement unknown-outcome reconciliation, late-success financial exception recording, and stuck-saga telemetry; verify neither uncertain capture nor contradictory result silently finalizes a workflow.

## 3. Orders command/result ownership

- [x] 3.1 Add idempotent checkout order-create command consumption in `services/orders/`, including immutable body comparison and a correlated outbox result in the same transaction; verify replay yields one order and conflict cannot mutate the original.
- [x] 3.2 Add explicit paid/failed finalization commands and correlated acknowledgements, rejecting conflicting terminal transitions; verify duplicates do not re-finalize and an order remains pending until the command arrives.
- [x] 3.3 Replace the old direct payment/reservation outcome listeners with Checkout command handling; verify only the new saga can finalize newly created checkout orders while existing production data remains untouched.

## 4. Inventory reservation and compensation

- [x] 4.1 Add correlated, idempotent reserve command/result processing in `services/inventory/` with all-or-nothing order line holds; verify duplicate commands do not double-reserve and insufficient stock leaves no partial hold.
- [x] 4.2 Add order-level cancellation tombstones that fence delayed reserve commands even when cancellation arrives first; verify the cancel-before-reserve race leaves stock available.
- [x] 4.3 Add idempotent commit/release commands and durable committed/released/absent results; verify a duplicate settlement cannot double-decrement or double-release quantity.
- [x] 4.4 Remove `orders.created` checkout auto-reserve and the old checkout reserve path; verify stock holds occur only after Checkout's reserve command.

## 5. Payment and hosted gateway

- [x] 5.1 Add Checkout-authorized session-create command consumption and correlated session-ready result in `services/payment/`, persisting one intent/transaction per order; verify retries reuse the identity and mismatched facts fail.
- [x] 5.2 Persist gateway callbacks and publish correlated definitive outcomes with idempotency; verify repeated, out-of-order, and contradictory callbacks cannot generate two effective charge outcomes.
- [x] 5.3 Add explicit cancellation/verification and uncertain-outcome reporting for payment deadlines, plus retained late-success exceptions; verify no stock release is ordered solely because a callback is late.
- [x] 5.4 Remove Payment's `inventory.reserved` catch-up consumer and route expiry through Checkout reconciliation; verify session creation needs no reservation event delivery to Payment.

## 6. Web buyer flow

- [x] 6.1 Replace sequential cart checkout mutation in `services/web/` with one snapshot submission to Checkout after existing batch product/merchant/shipping validation; verify no checkout POST calls Orders, Inventory, or Payment directly.
- [x] 6.2 Add a buyer-authenticated pending/progress page using the existing templ/Datastar style, redirecting only after session readiness; verify disconnect, refresh, and repeated submit reuse the original workflow.
- [x] 6.3 Keep gateway callback forwarding to Payment and render buyer-visible pending settlement on return; verify the browser never directly marks an order paid and cart removal honors the final checkout result.

## 7. Infrastructure, rollout, and documentation

- [x] 7.1 Add Checkout image/module to CI and Helm with CloudNativePG, migrations, resources, and telemetry configuration; verify Helm templates and CI image targets build/render.
- [x] 7.2 Add Checkout Debezium outbox connector, new topic routing, secrets, and mesh allow-lists; verify templates route only intended producers/consumers and web cannot directly mutate checkout inventory.
- [x] 7.3 Verify isolated production simulator checkouts for duplicate submission, stock shortage, payment failure, and payment success; confirm buyer-scoped status, order outcomes, and stock release/commit. Exercise lost acknowledgement, expiry, late success/manual review, and worker restart with isolated service tests instead of disruptive production fault injection; retain legacy data and buyer traffic.
- [x] 7.4 Verify the existing prod rollout in place: Checkout, its consumers/connectors and migrations are healthy; dashboard pending-age, compensation, and financial-exception counts, and confirm accepted sagas remain buyer-queryable without replacing legacy data.
- [x] 7.5 Remove superseded checkout RPC/consumer paths and unused old topics, and update `docs/order-placement.md` and diagrams; verify no old-path coordinator is deployed for new checkouts and docs match the in-place rollout.
