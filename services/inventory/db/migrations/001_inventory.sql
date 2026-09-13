-- +goose Up
CREATE TABLE IF NOT EXISTS inventory (
    product_id UUID PRIMARY KEY,
    available_qty INTEGER NOT NULL DEFAULT 0,
    reserved_qty INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_available_qty_non_negative CHECK (available_qty >= 0),
    CONSTRAINT inventory_reserved_qty_non_negative CHECK (reserved_qty >= 0)
);

-- +goose Down
DROP TABLE IF EXISTS inventory;
