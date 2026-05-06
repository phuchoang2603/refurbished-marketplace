-- name: UpsertPaymentIntent :one
INSERT INTO payment_intents (
    order_id,
    buyer_user_id,
    payment_token,
    currency,
    billing_address,
    shipping_address,
    status
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (order_id) DO UPDATE
SET buyer_user_id = excluded.buyer_user_id,
payment_token = excluded.payment_token,
currency = excluded.currency,
billing_address = excluded.billing_address,
shipping_address = excluded.shipping_address,
status = excluded.status,
updated_at = NOW()
RETURNING payment_intents.*;

-- name: GetPaymentIntentByOrderID :one
SELECT payment_intents.*
FROM payment_intents
WHERE order_id = $1;
