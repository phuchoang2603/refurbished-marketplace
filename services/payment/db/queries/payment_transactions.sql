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
    payment_transactions.id,
    payment_transactions.order_id,
    payment_transactions.merchant_id,
    payment_transactions.amount_cents,
    payment_transactions.currency,
    payment_transactions.status,
    payment_transactions.idempotency_key,
    payment_transactions.gateway_transaction_id,
    payment_transactions.failure_reason,
    payment_transactions.created_at,
    payment_transactions.updated_at;

-- name: GetPaymentTransactionByID :one
SELECT
    payment_transactions.id,
    payment_transactions.order_id,
    payment_transactions.merchant_id,
    payment_transactions.amount_cents,
    payment_transactions.currency,
    payment_transactions.status,
    payment_transactions.idempotency_key,
    payment_transactions.gateway_transaction_id,
    payment_transactions.failure_reason,
    payment_transactions.created_at,
    payment_transactions.updated_at
FROM payment_transactions
WHERE id = $1;

-- name: GetPaymentTransactionByOrderID :one
SELECT
    payment_transactions.id,
    payment_transactions.order_id,
    payment_transactions.merchant_id,
    payment_transactions.amount_cents,
    payment_transactions.currency,
    payment_transactions.status,
    payment_transactions.idempotency_key,
    payment_transactions.gateway_transaction_id,
    payment_transactions.failure_reason,
    payment_transactions.created_at,
    payment_transactions.updated_at
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
    payment_transactions.id,
    payment_transactions.order_id,
    payment_transactions.merchant_id,
    payment_transactions.amount_cents,
    payment_transactions.currency,
    payment_transactions.status,
    payment_transactions.idempotency_key,
    payment_transactions.gateway_transaction_id,
    payment_transactions.failure_reason,
    payment_transactions.created_at,
    payment_transactions.updated_at;
