-- name: InsertInventoryInboxMessage :one
INSERT INTO inventory_inbox (message_id)
VALUES ($1)
ON CONFLICT (message_id) DO NOTHING
RETURNING TRUE;

-- name: CreateInventoryOutbox :one
INSERT INTO inventory_outbox (
    id, aggregate_id, event_type, payload, tracingspancontext
)
VALUES ($1, $2, $3, $4, $5)
RETURNING inventory_outbox.*;
