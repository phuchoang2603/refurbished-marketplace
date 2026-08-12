-- name: CreateInventoryReservation :one
INSERT INTO inventory_reservations (order_id, product_id, quantity, status)
VALUES ($1, $2, $3, $4)
RETURNING
    inventory_reservations.order_id,
    inventory_reservations.product_id,
    inventory_reservations.quantity,
    inventory_reservations.status,
    inventory_reservations.created_at,
    inventory_reservations.updated_at;

-- name: ListActiveInventoryReservationsByOrderID :many
SELECT
    inventory_reservations.order_id,
    inventory_reservations.product_id,
    inventory_reservations.quantity,
    inventory_reservations.status,
    inventory_reservations.created_at,
    inventory_reservations.updated_at
FROM inventory_reservations
WHERE order_id = $1 AND status = 'RESERVED'
ORDER BY product_id
FOR UPDATE;

-- name: MarkInventoryReservationCommitted :one
UPDATE inventory_reservations
SET
    status = 'COMMITTED',
    updated_at = NOW()
WHERE
    order_id = $1
    AND product_id = $2
    AND status = 'RESERVED'
RETURNING
    inventory_reservations.order_id,
    inventory_reservations.product_id,
    inventory_reservations.quantity,
    inventory_reservations.status,
    inventory_reservations.created_at,
    inventory_reservations.updated_at;

-- name: MarkInventoryReservationReleased :one
UPDATE inventory_reservations
SET
    status = 'RELEASED',
    updated_at = NOW()
WHERE
    order_id = $1
    AND product_id = $2
    AND status = 'RESERVED'
RETURNING
    inventory_reservations.order_id,
    inventory_reservations.product_id,
    inventory_reservations.quantity,
    inventory_reservations.status,
    inventory_reservations.created_at,
    inventory_reservations.updated_at;
