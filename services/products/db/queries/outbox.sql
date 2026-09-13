-- name: CreateProductOutbox :exec
INSERT INTO products_outbox (
    id, aggregate_id, event_type, payload, tracingspancontext
)
VALUES ($1, $2, $3, $4, $5);
