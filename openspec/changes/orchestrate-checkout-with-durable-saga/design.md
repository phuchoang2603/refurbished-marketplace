## Context

See `proposal.md` for motivation and `specs/` for the behavior contract. The repository already uses Go services with service-local PostgreSQL, `sqlc`/goose, Kafka, Debezium outboxes, inbox records, idempotent order creation, and a browser-hosted payment simulator. Web currently blocks on Orders → Inventory gRPC → Payment gRPC; inventory can also reserve from `orders.created`, and payment's `inventory.reserved` consumer only catches up/acknowledges. The current simulator reports a result via web's callback, but a future payment gateway may have an uncertain or late capture.

## Goals / Non-Goals

**Goals:**

- Make one Checkout-owned, inspectable, durable state machine authoritative for checkout progression, stock compensation, timeouts, and finalization.
- Keep the stock-before-payment rule and one payment attempt per order, while allowing browser disconnects and message replay.
- Reuse service-local storage, event delivery, and observability patterns already in the repository.

**Non-Goals:**

- A cross-service transaction, exactly-once Kafka delivery, a synchronous guarantee that POST checkout reaches the gateway, or a generic saga framework.
- Moving cart state into Checkout, integrating a real provider, or migrating pre-existing checkout/order/payment data into the new flow.

## Decisions

### 1. Checkout owns the workflow, not the underlying order, stock, or money

Introduce `services/checkout/` with its own CloudNativePG database and status/submit gRPC API. Web validates the selected merchant group against Products, snapshots authoritative prices/names, authenticated buyer, merchant, shipping and return context, and submits a stable buyer-scoped intent; the service stores `(buyer_id, intent_key)` uniquely with a canonical request hash. Identical retries return the same checkout id, conflicting payloads fail. The response means accepted, not paid. Checkout is the only issuer of order-create, reserve, session-create, stock-settlement, and finalization commands for new checkouts. Orders, Inventory, and Payment retain their respective records and command validation.

Alternative: orchestrate inside Web or Orders. Web does not own durable workflow state and can disappear when the buyer disconnects; Orders would conflate business order truth with payment/stock coordination. Redis/Valkey is retained for cart state, not the authoritative saga ledger.

### 2. Dedicated PostgreSQL state, inbox, outbox, and due work in one transaction

Store `checkouts` (checkout id, buyer, merchant, request hash, immutable snapshot, order id if known, payment session id, state, version, deadline, last failure), `checkout_inbox` (unique message id), `checkout_outbox` (unique command/transition id), and a reconciliation/audit trail. A consumer locks the checkout row, verifies message identity/correlation and allowed transition, then writes inbox + state/version + outgoing command in one DB commit before acknowledging Kafka. Retries produce the original effective result; rejects and anomalies remain observable. Publish via the existing Debezium outbox pattern with `aggregate_id = checkout_id` so events for one checkout are keyed together. Each downstream service performs the same atomic inbox + domain write + outbox for checkout commands. Stable operation ids derive from checkout id and step, not delivery offset.

A deadline worker claims due rows with row locking and advances the same state machine; it does not delete state via a TTL. Suggested initial deadlines: short configurable order/reservation/session-preparation deadlines and a 30-minute payment deadline compatible with existing sessions. Timed-out uncertain commands stay pending reconciliation until they are definitively cancelled or observed. Transient settlement/finalization failures retry with backoff and alert, not silently finalize.

Alternative: Redis with TTLs and Kafka producer calls. It cannot commit the authoritative saga transition, inbox, and existing Postgres CDC outbox in the same local transaction and a TTL alone cannot prove compensation finished.

### 3. Explicit command/result contracts and state transitions

Use versioned protobuf contracts under `shared/proto/checkout/v1/` with checkout id, order id when known, message id, causation id, version/attempt and immutable commerce facts as needed. Use `checkout.*.v1` command/result Kafka topics with clear ownership, replacing the old payment/inventory-driven checkout topic flow. Orders/Inventory/Payment outbox rows use these event types for Checkout. Only Checkout consumes checkout results; Orders and Inventory no longer consume payment outcomes as independent checkout drivers.

Transition outline:

| State                               | Result                                       | Command/next state                        |
| ----------------------------------- | -------------------------------------------- | ----------------------------------------- |
| ACCEPTED                            | start committed                              | create order → CREATING_ORDER             |
| CREATING_ORDER                      | order created                                | reserve stock → RESERVING_STOCK           |
| RESERVING_STOCK                     | stock reserved                               | create hosted session → CREATING_SESSION  |
| CREATING_SESSION                    | session ready                                | expose payment URL → READY_TO_PAY         |
| READY_TO_PAY                        | definitive payment success                   | commit reservation → COMMITTING_STOCK     |
| COMMITTING_STOCK                    | stock committed                              | finalize order paid → FINALIZING_PAID     |
| READY_TO_PAY                        | definitive failure or confirmed cancellation | release reservation → RELEASING_STOCK     |
| RELEASING_STOCK                     | stock released/absent                        | finalize order failed → FINALIZING_FAILED |
| FINALIZING_PAID / FINALIZING_FAILED | matching order finalization                  | COMPLETED / FAILED                        |

An order-creation timeout with an unknown result is reconciled by idempotent retries; it cannot assume no order was written. A reservation timeout first issues a fencing cancellation. Inventory records a durable order-level cancelled tombstone even if no reservation exists; it rejects any later reserve for that order and reports absence/release. Session-creation timeout first asks Payment to cancel/confirm no capture before stock release. Every late result is checked against current state and the originating operation. No second session is created for an expired/terminal order; a new buyer attempt requires a new checkout/order id.

Alternative: reusing `orders.created` as an automatic Inventory reserve request and `inventory.reserved` as Payment's signal. This recreates the present competing coordinators and the missing-intent race.

### 4. Payment finality is a gate, not an optimistic timeout

Payment creates a session and transaction in one transaction after receiving Checkout's correlated reservation-authorized command. Its gateway callback ingress persists and emits definitive success/failure or uncertain outcome. For a payment deadline, Checkout requests cancellation/verification; Payment responds with cancelled/no-capture only when it can establish it. Until then the checkout remains non-terminal in a needs-reconciliation state, stock is not blindly released, and operators are alerted. A verified success arriving after definitive cancellation or order failure records a financial exception for investigation rather than issuing a second stock commit or silently discarding funds. Payment expiry and saga deadlines must not independently publish settlement commands. The simulator must model duplicate, delayed, and contradictory callbacks in tests; a future provider must add provider-specific signature verification, cancellation, reconciliation, and refunds/voids before real charging.

Alternative: unconditionally release at 30 minutes or interpret a missing callback as decline. Either could free stock when payment actually succeeded.

### 5. Buyer-facing progress and ownership boundaries

Web POST checkout calls only Checkout after local validation, then redirects to a buyer-authenticated progress URL. The page polls (or receives Datastar fragments) for status; only a READY_TO_PAY result can render/redirect to the gateway URL. On return, the order page shows processing until Checkout has observed settlement and finalization. Keep one-shot behavior and cart removal on buyer-visible terminal outcome as required by `specs/web/spec.md`. Checkout status reads check authenticated buyer identity server-side; a public session id alone is not authorization. Web forwards gateway callbacks only to Payment and never directly mutates saga state.

Alternative: block the HTTP request until Kafka finishes each step. This retains tail latency and couples buyer connection lifetime to asynchronous delivery; it is not fully event-driven.

### 6. Rollout on existing production and service ownership

Add Checkout image/workspace module, migrations/sqlc, Helm service/DB/jobs/secrets, a dedicated Debezium outbox connector, mesh rules for web → Checkout and gateway → web → Payment; remove web → Inventory checkout mutation permission. Preserve existing observability: correlate checkout/order ids and trace context through outboxes, add phase duration, pending-deadline, retry, compensation, reconciliation, and stuck-workflow metrics. Update `docs/order-placement.md` and architecture diagrams only when behavior ships, not as a substitute for this proposal.

Keep the existing production dataset and pending sessions intact. Replace the old web checkout sequence and event consumers for **new** checkouts rather than running two coordinators. Deploy Checkout storage, connectors, command handlers, and web flow together; check the new path before relying on it for buyer checkouts. Run failure scenarios only with isolated synthetic fixtures and without resetting shared databases, deleting legacy records, or interrupting buyer traffic. Once a checkout is accepted, preserve its saga state and complete or reconcile it under Checkout. Pre-existing legacy data is deliberately outside this change.

## Risks / Trade-offs

- [Extra Kafka hops delay the payment URL] → Show explicit pending progress, instrument p50/p95 by step, and avoid unnecessary round trips; this redesign does not solve slow payment DB commits automatically.
- [Payment success races with deadline/release] → Require definitive cancellation before release; retain unresolved money states and operator reconciliation rather than guessing.
- [Late reservation after timeout creates orphan stock] → Persist an Inventory cancellation tombstone that fences later reserve for the order.
- [Kafka redelivery or reordered results create duplicate side effects] → Stable command ids, local inbox/outbox transactions, order/session uniqueness, state/version guards and exception logging.
- [Incomplete rollout accepts new checkouts before consumers are ready] → Gate new checkout submissions on service, connector, and consumer readiness; verify one coordinator owns all new commands without touching legacy records.
- [User waits while services are unavailable] → Report durable progress, retry and alert on stuck workflows, and never create a second paid attempt under the same idempotency key.

## Migration Plan

1. Add contract/state-machine artifacts, Checkout DB and service, and replay/failure tests; no buyer traffic yet.
2. Replace the old Web reserve/session sequence, Inventory `orders.created` checkout auto-reserve, Payment `inventory.reserved` consumer, and direct outcome listeners on Orders/Inventory with the new command/result handlers.
3. Add the Web progress view and buyer-authenticated status API; migrate the existing production databases in place, bring up Checkout storage and connectors, and validate failure paths with isolated synthetic fixtures.
4. Verify all dependencies are healthy before relying on the new checkout flow; monitor pending-age, compensation, and financial-exception queues. Keep legacy data in place; there is no legacy checkout migration.
5. Never reset production data to repair a faulty rollout. Keep the new coordinator operating until accepted workflows finish or are reconciled; do not switch an in-progress saga to the removed old path.
