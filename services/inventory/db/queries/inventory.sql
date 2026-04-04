-- name: CreateInventory :one
INSERT INTO inventory (product_id, available_qty, reserved_qty)
VALUES ($1, $2, 0)
RETURNING inventory.*;

-- name: GetInventoryByProductID :one
SELECT inventory.*
FROM inventory
WHERE product_id = $1;

-- name: ReserveStock :one
UPDATE inventory
SET available_qty = available_qty - $2,
    reserved_qty = reserved_qty + $2,
    updated_at = NOW()
WHERE product_id = $1
  AND available_qty >= $2
RETURNING inventory.*;

-- name: CommitReservation :one
UPDATE inventory
SET reserved_qty = reserved_qty - $2,
    updated_at = NOW()
WHERE product_id = $1
  AND reserved_qty >= $2
RETURNING inventory.*;

-- name: ReleaseReservation :one
UPDATE inventory
SET available_qty = available_qty + $2,
    reserved_qty = reserved_qty - $2,
    updated_at = NOW()
WHERE product_id = $1
  AND reserved_qty >= $2
RETURNING inventory.*;
