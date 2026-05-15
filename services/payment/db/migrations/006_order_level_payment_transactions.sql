-- +goose Up
DROP INDEX IF EXISTS payment_transactions_order_item_id_uniq;

ALTER TABLE payment_transactions
DROP COLUMN IF EXISTS order_item_id;

CREATE UNIQUE INDEX IF NOT EXISTS payment_transactions_order_id_uniq ON payment_transactions (
    order_id
);

-- +goose Down
DROP INDEX IF EXISTS payment_transactions_order_id_uniq;

ALTER TABLE payment_transactions
ADD COLUMN IF NOT EXISTS order_item_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

CREATE UNIQUE INDEX IF NOT EXISTS payment_transactions_order_item_id_uniq ON payment_transactions (
    order_item_id
);
