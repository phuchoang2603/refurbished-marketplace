-- name: CreatePaymentTransaction :one
INSERT INTO payment_transactions (
    id,
    order_id,
    merchant_id,
    amount_cents,
    currency,
    status,
    idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (order_id) DO NOTHING
RETURNING
    payment_transactions.*;

-- name: GetPaymentTransactionByID :one
SELECT *
FROM payment_transactions
WHERE id = $1;

-- name: GetPaymentTransactionByOrderID :one
SELECT *
FROM payment_transactions
WHERE order_id = $1;

-- name: UpdatePaymentTransactionGatewayResult :one
UPDATE payment_transactions
SET
    status = $2,
    gateway_transaction_id = $3,
    failure_reason = $4,
    updated_at = NOW()
WHERE id = $1 AND status NOT IN ('SUCCEEDED', 'FAILED')
RETURNING
    payment_transactions.*;
