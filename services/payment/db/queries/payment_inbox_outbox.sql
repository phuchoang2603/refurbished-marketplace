-- name: InsertPaymentInboxMessage :one
INSERT INTO payment_inbox (message_id)
VALUES ($1)
ON CONFLICT (message_id) DO NOTHING
RETURNING TRUE;

-- name: CreatePaymentOutbox :one
INSERT INTO payment_outbox (
    id, aggregate_id, event_type, payload, tracingspancontext
)
VALUES ($1, $2, $3, $4, $5)
RETURNING
    payment_outbox.id,
    payment_outbox.aggregate_id,
    payment_outbox.event_type,
    payment_outbox.payload,
    payment_outbox.publish_attempts,
    payment_outbox.created_at,
    payment_outbox.published_at,
    payment_outbox.tracingspancontext;

-- name: ListPaymentOutboxByAggregateID :many
SELECT
    payment_outbox.id,
    payment_outbox.aggregate_id,
    payment_outbox.event_type,
    payment_outbox.payload,
    payment_outbox.publish_attempts,
    payment_outbox.created_at,
    payment_outbox.published_at,
    payment_outbox.tracingspancontext
FROM payment_outbox
WHERE aggregate_id = $1
ORDER BY created_at;
