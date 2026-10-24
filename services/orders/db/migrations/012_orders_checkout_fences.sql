-- +goose Up
ALTER TABLE orders_checkout_commands
DROP CONSTRAINT IF EXISTS orders_checkout_commands_order_id_fkey;
ALTER TABLE orders_checkout_commands ALTER COLUMN request_hash DROP NOT NULL;
ALTER TABLE orders_checkout_commands ADD COLUMN status TEXT NOT NULL DEFAULT 'CREATED' CHECK (
    status IN ('CREATED', 'REJECTED', 'FAILED_ABSENT')
);
ALTER TABLE orders_checkout_commands ADD COLUMN reason TEXT;

-- +goose Down
DELETE FROM orders_checkout_commands WHERE status <> 'CREATED';
ALTER TABLE orders_checkout_commands DROP COLUMN reason;
ALTER TABLE orders_checkout_commands DROP COLUMN status;
ALTER TABLE orders_checkout_commands ALTER COLUMN request_hash SET NOT NULL;
ALTER TABLE orders_checkout_commands
ADD CONSTRAINT orders_checkout_commands_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES orders(id);
