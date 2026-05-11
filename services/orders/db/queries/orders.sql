-- name: CreateOrder :one
INSERT INTO orders (id, buyer_user_id, status, total_cents)
VALUES ($1, $2, $3, $4)
RETURNING id, buyer_user_id, status, total_cents, created_at, updated_at;

-- name: CreateOrderOutbox :one
INSERT INTO orders_outbox (id, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4)
RETURNING
    orders_outbox.id, orders_outbox.aggregate_id, orders_outbox.event_type, orders_outbox.payload, orders_outbox.publish_attempts, orders_outbox.created_at, orders_outbox.published_at;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    id,
    order_id,
    product_id,
    merchant_id,
    quantity,
    unit_price_cents,
    line_total_cents
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    id, order_id, product_id, quantity, unit_price_cents, line_total_cents, created_at, merchant_id;

-- name: GetOrderByID :one
SELECT id, buyer_user_id, status, total_cents, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: ListOrdersByBuyer :many
SELECT id, buyer_user_id, status, total_cents, created_at, updated_at
FROM orders
WHERE buyer_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateOrderStatus :one
UPDATE orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, buyer_user_id, status, total_cents, created_at, updated_at;

-- name: ListOrderItemsByOrderIDs :many
SELECT
    id,
    order_id,
    product_id,
    quantity,
    unit_price_cents,
    line_total_cents,
    created_at,
    merchant_id
FROM order_items
WHERE order_id = ANY($1::uuid [])
ORDER BY order_id ASC, created_at ASC;
