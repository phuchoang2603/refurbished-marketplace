-- +goose Up
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES payment_intents(order_id) ON DELETE CASCADE,
    order_item_id UUID NOT NULL,
    merchant_id UUID NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    status TEXT NOT NULL DEFAULT 'INITIALIZED',
    idempotency_key TEXT NOT NULL,
    gateway_transaction_id TEXT,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS payment_transactions_order_item_id_uniq ON payment_transactions (order_item_id);
CREATE UNIQUE INDEX IF NOT EXISTS payment_transactions_idempotency_key_uniq ON payment_transactions (idempotency_key);
CREATE INDEX IF NOT EXISTS payment_transactions_order_id_idx ON payment_transactions (order_id);
CREATE INDEX IF NOT EXISTS payment_transactions_status_idx ON payment_transactions (status);

-- +goose Down
DROP TABLE IF EXISTS payment_transactions;

