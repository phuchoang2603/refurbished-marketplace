-- name: InsertCheckout :one
INSERT INTO checkouts (
    id,
    buyer_user_id,
    merchant_id,
    intent_key,
    request_hash,
    snapshot,
    state,
    order_id,
    deadline_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
    payment_session_id
    = COALESCE(sqlc.narg(payment_session_id)::text, payment_session_id),
    payment_return_url
    = COALESCE(sqlc.narg(payment_return_url)::text, payment_return_url),
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
