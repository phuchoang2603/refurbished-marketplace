-- name: InsertInbox :execrows
INSERT INTO checkout_inbox (message_id, checkout_id)
VALUES ($1, $2)
ON CONFLICT (message_id) DO NOTHING;

-- name: InsertOutbox :exec
INSERT INTO checkout_outbox (
    id, aggregate_id, event_type, payload, tracingspancontext
)
VALUES ($1, $2, $3, $4, $5);

-- name: InsertException :exec
INSERT INTO checkout_exceptions (id, checkout_id, message_id, kind, details)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (checkout_id, message_id, kind) DO NOTHING;

-- name: GetLatestOutbox :one
SELECT *
FROM checkout_outbox
WHERE aggregate_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 1
;
