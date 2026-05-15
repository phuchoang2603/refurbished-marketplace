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

-- name: ListOrderItemsByOrderIDs :many
SELECT *
FROM order_items
WHERE order_id = ANY($1::uuid [])
ORDER BY order_id ASC, created_at ASC;
