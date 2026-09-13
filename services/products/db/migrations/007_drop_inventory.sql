-- +goose Up
ALTER TABLE inventory_reservations
DROP CONSTRAINT IF EXISTS inventory_reservations_product_id_fk;
ALTER TABLE inventory
DROP CONSTRAINT IF EXISTS inventory_product_id_fk;
DROP TABLE IF EXISTS inventory_reservations;
DROP TABLE IF EXISTS inventory_inbox;
DROP TABLE IF EXISTS inventory_outbox;
DROP TABLE IF EXISTS inventory;

-- +goose Down
SELECT 1;
