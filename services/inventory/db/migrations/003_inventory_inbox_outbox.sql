-- +goose Up
CREATE TABLE IF NOT EXISTS inventory_inbox (
    message_id TEXT PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    publish_attempts INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS inventory_outbox_aggregate_id_idx ON inventory_outbox (
    aggregate_id
);
CREATE INDEX IF NOT EXISTS inventory_outbox_event_type_idx ON inventory_outbox (
    event_type
);
CREATE INDEX IF NOT EXISTS inventory_outbox_published_at_idx ON inventory_outbox (
    published_at
);

-- +goose Down
DROP TABLE IF EXISTS inventory_outbox;
DROP TABLE IF EXISTS inventory_inbox;
