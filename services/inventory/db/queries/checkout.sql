-- name: GetCheckoutReservation :one
SELECT * FROM inventory_checkout_operations WHERE order_id = $1;

-- name: InsertCheckoutReservation :exec
INSERT INTO inventory_checkout_operations (
    order_id, checkout_id, reserve_hash, status
)
VALUES ($1, $2, $3, $4);

-- name: SetCheckoutReservationStatus :execrows
UPDATE inventory_checkout_operations SET status = $2, updated_at = NOW()
WHERE order_id = $1 AND status = $3;

-- name: GetOrderReservations :many
SELECT *
FROM inventory_reservations
WHERE order_id = $1
ORDER BY product_id FOR UPDATE;
