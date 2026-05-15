-- name: CreateOrder :one
INSERT INTO orders (id, buyer_user_id, merchant_id, status, total_cents)
VALUES ($1, $2, $3, $4, $5)
RETURNING
    orders.*;

-- name: CreateOrderOutbox :one
INSERT INTO orders_outbox (id, aggregate_id, event_type, payload)
VALUES ($1, $2, $3, $4)
RETURNING
    orders_outbox.*;

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
