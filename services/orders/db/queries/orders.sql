-- name: CreateOrder :one
INSERT INTO orders (id, buyer_user_id, product_id, quantity, status, total_cents)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING orders.*;

-- name: GetOrderByID :one
SELECT orders.*
FROM orders
WHERE id = $1;

-- name: ListOrdersByBuyer :many
SELECT orders.*
FROM orders
WHERE buyer_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING orders.*;
