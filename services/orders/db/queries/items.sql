-- name: CreateOrderItem :one
INSERT INTO order_items (
    id,
    order_id,
    product_id,
    quantity,
    unit_price_cents,
    line_total_cents
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    order_items.*;

-- name: ListOrderItemsByOrderIDs :many
SELECT *
FROM order_items
WHERE order_id = ANY($1::uuid [])
ORDER BY order_id ASC, created_at ASC;
