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
    payment_outbox.*;

-- name: ListPaymentOutboxByAggregateID :many
SELECT *
FROM payment_outbox
WHERE aggregate_id = $1
ORDER BY created_at;
