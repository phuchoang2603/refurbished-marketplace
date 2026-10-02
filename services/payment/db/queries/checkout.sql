-- name: GetCheckoutPaymentSession :one
SELECT * FROM payment_checkout_sessions WHERE order_id = $1;

-- name: InsertCheckoutPaymentSession :exec
INSERT INTO payment_checkout_sessions (
    order_id,
    checkout_id,
    request_hash,
    create_operation_id,
    create_version,
    status
)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertCheckoutCallback :execrows
INSERT INTO payment_checkout_callbacks (order_id, outcome)
VALUES ($1, $2) ON CONFLICT (order_id, outcome) DO NOTHING;

-- name: CountCheckoutCallbacks :one
SELECT count(*) FROM payment_checkout_callbacks WHERE order_id = $1;
