-- +goose Up
CREATE TABLE IF NOT EXISTS payment_intents (
    order_id UUID PRIMARY KEY,
    buyer_user_id UUID NOT NULL,
    payment_token TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    billing_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    shipping_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'INITIALIZED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS payment_intents_buyer_user_id_idx ON payment_intents (buyer_user_id);

-- +goose Down
DROP TABLE IF EXISTS payment_intents;

