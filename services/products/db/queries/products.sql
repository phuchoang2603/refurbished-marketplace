-- name: CreateListing :exec
INSERT INTO products (
    id,
    name,
    description,
    category,
    condition,
    price_cents,
    merchant_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: InsertOutboxEvent :exec
INSERT INTO products_outbox (
    id, aggregate_id, event_type, payload, tracingspancontext
)
VALUES ($1, $2, $3, $4, $5);

-- name: GetListingByID :one
SELECT * FROM products WHERE id = $1;

-- name: GetListingsByIDs :many
SELECT * FROM products WHERE id = ANY($1::uuid []);
