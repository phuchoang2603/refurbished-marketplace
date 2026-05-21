-- name: CreateInventory :one
INSERT INTO inventory (product_id, available_qty, reserved_qty)
VALUES ($1, $2, 0)
RETURNING inventory.*;

-- name: GetInventoryByProductID :one
SELECT *
FROM inventory
WHERE product_id = $1;

-- name: GetInventoriesByProductIDsForUpdate :many
SELECT *
FROM inventory
WHERE product_id = ANY($1::uuid [])
FOR UPDATE;

-- name: ReserveInventoryStock :one
UPDATE inventory
SET
    available_qty = available_qty - sqlc.arg(quantity),
    reserved_qty = reserved_qty + sqlc.arg(quantity),
    updated_at = NOW()
WHERE product_id = sqlc.arg(product_id)
RETURNING inventory.*;

-- name: CommitInventoryReservedStock :one
UPDATE inventory
SET
    reserved_qty = reserved_qty - sqlc.arg(quantity),
    updated_at = NOW()
WHERE
    product_id = sqlc.arg(product_id)
    AND reserved_qty >= sqlc.arg(quantity)
RETURNING inventory.*;

-- name: ReleaseInventoryReservedStock :one
UPDATE inventory
SET
    available_qty = available_qty + sqlc.arg(quantity),
    reserved_qty = reserved_qty - sqlc.arg(quantity),
    updated_at = NOW()
WHERE
    product_id = sqlc.arg(product_id)
    AND reserved_qty >= sqlc.arg(quantity)
RETURNING inventory.*;
