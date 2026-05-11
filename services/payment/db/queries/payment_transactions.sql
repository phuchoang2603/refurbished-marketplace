-- name: CreatePaymentTransaction :one
INSERT INTO payment_transactions (
    id,
    order_id,
    order_item_id,
    merchant_id,
    amount_cents,
    currency,
    status,
    idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (order_item_id) DO NOTHING
RETURNING
    id, order_id, order_item_id, merchant_id, amount_cents, currency, status, idempotency_key, gateway_transaction_id, failure_reason, created_at, updated_at;

-- name: GetPaymentTransactionByID :one
SELECT
    id,
    order_id,
    order_item_id,
    merchant_id,
    amount_cents,
    currency,
    status,
    idempotency_key,
    gateway_transaction_id,
    failure_reason,
    created_at,
    updated_at
FROM payment_transactions
WHERE id = $1;

-- name: GetPaymentTransactionByOrderItemID :one
SELECT
    id,
    order_id,
    order_item_id,
    merchant_id,
    amount_cents,
    currency,
    status,
    idempotency_key,
    gateway_transaction_id,
    failure_reason,
    created_at,
    updated_at
FROM payment_transactions
WHERE order_item_id = $1;

-- name: UpdatePaymentTransactionGatewayResult :one
UPDATE payment_transactions
SET
    status = $2,
    gateway_transaction_id = $3,
    failure_reason = $4,
    updated_at = NOW()
WHERE id = $1 AND status NOT IN ('SUCCEEDED', 'FAILED')
RETURNING
    id, order_id, order_item_id, merchant_id, amount_cents, currency, status, idempotency_key, gateway_transaction_id, failure_reason, created_at, updated_at;
