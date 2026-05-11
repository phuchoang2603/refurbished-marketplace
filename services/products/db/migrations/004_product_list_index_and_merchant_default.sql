-- +goose Up
ALTER TABLE products
ALTER COLUMN merchant_id DROP DEFAULT;

CREATE INDEX IF NOT EXISTS products_created_at_id_idx ON products (
    created_at DESC, id DESC
);

-- +goose Down
DROP INDEX IF EXISTS products_created_at_id_idx;

ALTER TABLE products
ALTER COLUMN merchant_id SET DEFAULT '00000000-0000-0000-0000-000000000000';
