-- name: GetOrderByID :one
SELECT *
FROM orders
WHERE id = $1
LIMIT 1;

-- name: ListOrdersByBuyer :many
SELECT *
FROM orders
WHERE buyer_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrderByBuyerIdempotencyKey :one
SELECT *
FROM orders
WHERE buyer_user_id = $1 AND idempotency_key = $2
LIMIT 1;

-- name: CreateOrder :one
INSERT INTO orders (
    id,
    buyer_user_id,
    merchant_id,
    status,
    total_cents,
    idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    orders.*;

-- name: UpdateOrderStatus :one
UPDATE orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    orders.*;
