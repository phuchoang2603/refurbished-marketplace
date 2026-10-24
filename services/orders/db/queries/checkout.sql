-- name: GetCheckoutOrderCommand :one
SELECT * FROM orders_checkout_commands WHERE checkout_id = $1;

-- name: IsCheckoutOrder :one
SELECT EXISTS(SELECT 1 FROM orders_checkout_commands WHERE order_id = $1);

-- name: LockCheckoutIdentity :exec
SELECT PG_ADVISORY_XACT_LOCK(HASHTEXTEXTENDED($1::text, 0));

-- name: InsertCheckoutOrderCommand :exec
INSERT INTO orders_checkout_commands (
    checkout_id, order_id, request_hash, status, reason
)
VALUES ($1, $2, $3, $4, $5);

-- name: FinalizeCheckoutOrder :one
UPDATE orders SET status = $2, updated_at = NOW()
WHERE id = $1 AND status IN ('ORDER_STATUS_PENDING', $2)
RETURNING *;
