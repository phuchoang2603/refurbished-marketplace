-- +goose Up
DROP INDEX IF EXISTS orders_buyer_user_id_idx;
DROP INDEX IF EXISTS order_items_order_id_idx;

CREATE INDEX IF NOT EXISTS orders_buyer_user_id_created_at_idx ON orders (
    buyer_user_id, created_at DESC
);
CREATE INDEX IF NOT EXISTS order_items_order_id_created_at_idx ON order_items (
    order_id, created_at
);

-- +goose Down
DROP INDEX IF EXISTS order_items_order_id_created_at_idx;
DROP INDEX IF EXISTS orders_buyer_user_id_created_at_idx;

CREATE INDEX IF NOT EXISTS orders_buyer_user_id_idx ON orders (buyer_user_id);
CREATE INDEX IF NOT EXISTS order_items_order_id_idx ON order_items (order_id);
