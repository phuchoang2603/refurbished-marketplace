-- name: CreateOrderOutbox :one
INSERT INTO orders_outbox (id, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4)
RETURNING
    orders_outbox.*;
