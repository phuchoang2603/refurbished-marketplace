-- name: InsertCheckout :one
INSERT INTO checkouts (
    id,
    buyer_user_id,
    merchant_id,
    intent_key,
    request_hash,
    snapshot,
    gateway_snapshot,
    state,
    order_id,
    deadline_at,
    prep_operation_id,
    prep_version,
    prep_deadline_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (buyer_user_id, intent_key) DO NOTHING
RETURNING *;

-- name: GetCheckout :one
SELECT * FROM checkouts WHERE id = $1;

-- name: GetCheckoutByIntent :one
SELECT * FROM checkouts WHERE buyer_user_id = $1 AND intent_key = $2;

-- name: LockCheckout :one
SELECT * FROM checkouts WHERE id = $1 FOR UPDATE;

-- name: AdvanceCheckout :execrows
UPDATE checkouts
SET
    state = $2,
    version = version + 1,
    order_id = COALESCE(sqlc.narg(order_id)::uuid, order_id),
    gateway_payment_id
    = COALESCE(sqlc.narg(gateway_payment_id)::text, gateway_payment_id),
    hosted_checkout_url
    = COALESCE(sqlc.narg(hosted_checkout_url)::text, hosted_checkout_url),
    failure_reason = COALESCE(sqlc.narg(failure_reason)::text, failure_reason),
    deadline_at = sqlc.narg(deadline_at)::timestamptz, updated_at = NOW()
WHERE id = $1 AND version = $3;

-- name: ClaimDueCheckouts :many
SELECT * FROM checkouts
WHERE deadline_at <= NOW()
ORDER BY deadline_at, id
LIMIT $1 FOR UPDATE SKIP LOCKED;

-- name: DelayCheckoutDeadline :execrows
UPDATE checkouts SET deadline_at = $2, updated_at = NOW()
WHERE id = $1 AND version = $3;

-- name: GetCheckoutByOrder :one
SELECT * FROM checkouts WHERE order_id = $1;

-- name: SetPreparationOutcome :exec
UPDATE checkouts
SET
    orders_outcome = CASE
        WHEN sqlc.arg(participant)::text = 'orders'
            THEN sqlc.arg(outcome)::text
        ELSE orders_outcome
    END,
    orders_reason = CASE
        WHEN sqlc.arg(participant)::text = 'orders'
            THEN sqlc.narg(reason)::text
        ELSE orders_reason
    END,
    inventory_outcome = CASE
        WHEN sqlc.arg(participant)::text = 'inventory'
            THEN sqlc.arg(outcome)::text
        ELSE inventory_outcome
    END,
    inventory_reason = CASE
        WHEN sqlc.arg(participant)::text = 'inventory'
            THEN sqlc.narg(reason)::text
        ELSE inventory_reason
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: ClosePreparation :execrows
UPDATE checkouts
SET prep_closure = $2, prep_closed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND prep_closed_at IS NULL;
