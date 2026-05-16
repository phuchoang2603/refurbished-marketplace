## 1. Shared contracts

- [x] 1.1 Extend `shared/proto/orders/v1/order_created.proto` to include order item lines required for reservation.
- [x] 1.2 Add shared protobuf contracts for `inventory.reserved` and `inventory.reservation_failed` event payloads.
- [x] 1.3 Regenerate protobuf outputs and update shared messaging constants/topic wiring for the new inventory reservation outcomes.

## 2. Inventory reservation ownership

- [x] 2.1 Add inventory persistence for order-owned reservation records and any required inbox/outbox tables under `services/inventory/db/`.
- [x] 2.2 Generate updated sqlc queries and implement inventory service methods for idempotent order reservation, commit, and release using the reservation ledger.
- [x] 2.3 Add Kafka consumer and producer paths in `services/inventory` for enriched `orders.created`, `payment.succeeded` / `payment.failed`, and inventory reservation outcome events.
- [x] 2.4 Add inventory tests covering successful multi-line reservation, duplicate message handling, and rollback on partial reservation failure.

## 3. Orders and payment sequencing

- [x] 3.1 Update `services/orders` to publish enriched `orders.created` payloads with item lines.
- [x] 3.2 Update `services/payment` to consume `inventory.reserved` instead of `orders.created` and to ignore failed reservation paths.
- [x] 3.3 Update `services/orders` to consume `inventory.reservation_failed` and transition affected orders to the failed state.
- [x] 3.4 Add or update orders and payment tests for the revised event sequence and deduplication behavior.

## 4. Flow validation and docs

- [x] 4.1 Update `docs/core-order-payment-flow.md` to reflect the inventory reservation step and revised Kafka choreography.
- [x] 4.2 Run targeted service tests for `orders`, `inventory`, and `payment`, and fix any contract or migration issues uncovered by the new flow.
