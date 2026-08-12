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
    order_items.id,
    order_items.order_id,
    order_items.product_id,
    order_items.quantity,
    order_items.unit_price_cents,
    order_items.line_total_cents,
    order_items.created_at;

-- name: ListOrderItemsByOrderIDs :many
SELECT
    order_items.id,
    order_items.order_id,
    order_items.product_id,
    order_items.quantity,
    order_items.unit_price_cents,
    order_items.line_total_cents,
    order_items.created_at
FROM order_items
WHERE order_id = ANY($1::uuid [])
ORDER BY order_id ASC, created_at ASC;
