-- +goose Up
CREATE TABLE IF NOT EXISTS inventory_inbox (
    message_id TEXT PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS inventory_inbox;
