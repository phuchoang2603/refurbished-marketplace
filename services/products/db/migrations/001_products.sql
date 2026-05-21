-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price_cents BIGINT NOT NULL,
    merchant_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS products_created_at_id_idx ON products (
    created_at DESC, id DESC
);

-- +goose Down
DROP INDEX IF EXISTS products_created_at_id_idx;
DROP TABLE IF EXISTS products;
