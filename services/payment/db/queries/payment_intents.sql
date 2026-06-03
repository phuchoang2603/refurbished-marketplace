-- name: CreateHostedPaymentSession :one
INSERT INTO payment_intents (
    order_id,
    buyer_user_id,
    currency,
    shipping_address,
    status,
    payment_session_id,
    return_url,
    cancel_url,
    expires_at,
    failure_reason
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING payment_intents.*;

-- name: GetPaymentIntentByOrderID :one
SELECT *
FROM payment_intents
WHERE order_id = $1;

-- name: UpdateHostedPaymentSessionOutcome :one
UPDATE payment_intents
SET
    status = $3,
    failure_reason = $4,
    updated_at = NOW()
WHERE order_id = $1 AND payment_session_id = $2
RETURNING payment_intents.*;
