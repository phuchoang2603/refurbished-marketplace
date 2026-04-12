-- name: InsertPaymentInboxMessage :exec
INSERT INTO payment_inbox (message_id)
VALUES ($1)
ON CONFLICT (message_id) DO NOTHING;

-- name: CreatePaymentOutbox :one
INSERT INTO payment_outbox (id, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4)
RETURNING payment_outbox.id, payment_outbox.aggregate_id, payment_outbox.event_type, payment_outbox.payload, payment_outbox.publish_attempts, payment_outbox.created_at, payment_outbox.published_at;

-- name: GetPaymentOutboxByAggregateIDAndEventType :one
SELECT id, aggregate_id, event_type, payload, publish_attempts, created_at, published_at
FROM payment_outbox
WHERE aggregate_id = $1 AND event_type = $2
LIMIT 1;
