-- +goose Up
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS owner_user_id UUID;

UPDATE products
SET owner_user_id = gen_random_uuid()
WHERE owner_user_id IS NULL;

ALTER TABLE products
    ALTER COLUMN owner_user_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS products_owner_user_id_idx ON products (owner_user_id);

-- +goose Down
DROP INDEX IF EXISTS products_owner_user_id_idx;

ALTER TABLE products
    DROP COLUMN IF EXISTS owner_user_id;
