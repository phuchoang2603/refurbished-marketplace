-- name: InsertOrdersInboxMessage :one
INSERT INTO orders_inbox (message_id)
VALUES ($1)
ON CONFLICT (message_id) DO NOTHING
RETURNING TRUE;

-- name: CreateOrderOutbox :one
INSERT INTO orders_outbox (id, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4)
RETURNING
    orders_outbox.*;
