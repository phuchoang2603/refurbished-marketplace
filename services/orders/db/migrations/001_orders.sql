-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    buyer_user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    status TEXT NOT NULL,
    total_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_buyer_user_id_idx ON orders (buyer_user_id);
CREATE INDEX IF NOT EXISTS orders_product_id_idx ON orders (product_id);
CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);

-- +goose Down
DROP TABLE IF EXISTS orders;
