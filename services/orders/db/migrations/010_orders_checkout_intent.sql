-- +goose Up
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS idempotency_key UUID;

UPDATE orders
SET idempotency_key = id
WHERE idempotency_key IS NULL;

ALTER TABLE orders
ALTER COLUMN idempotency_key SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS orders_buyer_idempotency_key_uidx ON orders (
    buyer_user_id, idempotency_key
);

-- +goose Down
DROP INDEX IF EXISTS orders_buyer_idempotency_key_uidx;
ALTER TABLE orders
DROP COLUMN IF EXISTS idempotency_key;
