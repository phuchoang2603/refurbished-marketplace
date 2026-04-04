# Eventing Plan

## Goal

Use Kafka as the async backbone, but make event publishing reliable with outbox/inbox patterns.

## Outbox

- `orders` and later `payment` should write their business row and an outbox row in the same DB transaction.
- The outbox row is the durable event source of truth.
- Keep the outbox table local to the service that owns the business change.

## CDC

- Use Debezium to stream outbox rows from Postgres into Kafka.
- Prefer CDC over custom polling for low-latency and fewer moving parts.
- Kafka topics should mirror domain event families, not application internals.

## Inbox

- Consumer services should store processed message IDs in a local inbox table.
- This prevents duplicate processing when Kafka redelivers a message.
- `inventory` and `payment` are likely consumers to need inbox dedupe.

## Flow

1. `web` calls `orders`.
2. `orders` stores the order and an outbox event in one transaction.
3. Debezium streams the outbox row into Kafka.
4. `inventory` consumes the event and reserves stock.
5. `payment` consumes the event and checks its inbox table before processing.
6. `payment` records success/failure and emits follow-up events as needed.

## Why This Matters

- Prevents lost events when a service crashes after a DB write.
- Prevents double-processing on consumer retries.
- Keeps fraud/analytics consumers aligned with the source of truth.
