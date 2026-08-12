-- name: GetOrderByID :one
SELECT
    orders.id,
    orders.buyer_user_id,
    orders.status,
    orders.total_cents,
    orders.created_at,
    orders.updated_at,
    orders.merchant_id
FROM orders
WHERE id = $1
LIMIT 1;

-- name: ListOrdersByBuyer :many
SELECT
    orders.id,
    orders.buyer_user_id,
    orders.status,
    orders.total_cents,
    orders.created_at,
    orders.updated_at,
    orders.merchant_id
FROM orders
WHERE buyer_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrderByBuyerAndCheckoutIntentKey :one
SELECT
    orders.id,
    orders.buyer_user_id,
    orders.status,
    orders.total_cents,
    orders.created_at,
    orders.updated_at,
    orders.merchant_id
FROM orders
WHERE
    buyer_user_id = $1
    AND checkout_intent_idempotency_key = $2
LIMIT 1;

-- name: CreateOrder :one
INSERT INTO orders (
    id,
    buyer_user_id,
    checkout_intent_idempotency_key,
    merchant_id,
    status,
    total_cents
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    orders.id,
    orders.buyer_user_id,
    orders.status,
    orders.total_cents,
    orders.created_at,
    orders.updated_at,
    orders.merchant_id;

-- name: UpdateOrderStatus :one
UPDATE orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    orders.id,
    orders.buyer_user_id,
    orders.status,
    orders.total_cents,
    orders.created_at,
    orders.updated_at,
    orders.merchant_id;
