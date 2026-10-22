-- name: InsertGatewayOperation :exec
INSERT INTO checkout_gateway_operations (
    checkout_id, request, tracingspancontext
) VALUES ($1, $2, $3);

-- name: ClaimGatewayOperation :one
UPDATE checkout_gateway_operations
SET attempts = attempts + 1, next_attempt_at = NOW() + INTERVAL '30 seconds'
WHERE
    checkout_id = (
        SELECT checkout_id FROM checkout_gateway_operations
        WHERE status = 'pending' AND next_attempt_at <= NOW()
        ORDER BY next_attempt_at, checkout_id
        LIMIT 1 FOR UPDATE SKIP LOCKED
    )
RETURNING *;

-- name: FinishGatewayOperation :exec
UPDATE checkout_gateway_operations SET status = $2
WHERE checkout_id = $1 AND status = 'pending'
;

-- name: RetryGatewayOperation :exec
UPDATE checkout_gateway_operations SET next_attempt_at = $2
WHERE checkout_id = $1 AND status = 'pending' AND attempts = $3
;

-- name: RecordGatewayPayment :exec
UPDATE checkouts SET
    gateway_payment_id = $2,
    hosted_checkout_url = $3,
    gateway_expires_at = $4,
    payment_status = 'OPEN'
WHERE id = $1;

-- name: LockCheckoutByGatewayPayment :one
SELECT * FROM checkouts WHERE gateway_payment_id = $1 FOR UPDATE;

-- name: RecordGatewayEvent :exec
UPDATE checkouts SET
    gateway_last_event_at = GREATEST(gateway_last_event_at, $2::TIMESTAMPTZ),
    payment_status = COALESCE(sqlc.narg(payment_status)::TEXT, payment_status)
WHERE id = $1;

-- name: MarkGatewayStatusChecked :exec
UPDATE checkouts SET gateway_status_checked_at = NOW(), deadline_at = NULL
WHERE id = $1
;

-- name: ResolvePaymentUncertain :exec
UPDATE checkout_exceptions SET resolved_at = NOW()
WHERE checkout_id = $1 AND kind = 'payment_uncertain' AND resolved_at IS NULL
;
