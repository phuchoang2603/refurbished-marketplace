-- +goose Up
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS checkout_intent_idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS orders_buyer_checkout_intent_idx ON orders (
    buyer_user_id,
    checkout_intent_idempotency_key
) WHERE checkout_intent_idempotency_key IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS orders_buyer_checkout_intent_idx;

ALTER TABLE orders
DROP COLUMN IF EXISTS checkout_intent_idempotency_key;
