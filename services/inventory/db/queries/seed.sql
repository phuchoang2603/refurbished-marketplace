-- name: InsertInventorySeedIntent :exec
INSERT INTO inventory_seed_intents (product_id, initial_qty, product_version)
VALUES ($1, $2, $3)
ON CONFLICT (product_id) DO NOTHING;

-- name: GetInventorySeedIntent :one
SELECT product_id, initial_qty, product_version FROM inventory_seed_intents
WHERE product_id = $1 FOR UPDATE;
