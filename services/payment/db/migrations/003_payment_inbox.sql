-- +goose Up
CREATE TABLE IF NOT EXISTS payment_inbox (
    message_id TEXT PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS payment_inbox;

