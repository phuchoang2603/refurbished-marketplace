-- name: UpdateOrderStatus :one
UPDATE orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    orders.*;
