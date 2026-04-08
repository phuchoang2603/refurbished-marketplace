-- +goose Up
CREATE TABLE IF NOT EXISTS orders_outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    publish_attempts INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS orders_outbox_aggregate_id_idx ON orders_outbox (aggregate_id);
CREATE INDEX IF NOT EXISTS orders_outbox_event_type_idx ON orders_outbox (event_type);
CREATE INDEX IF NOT EXISTS orders_outbox_published_at_idx ON orders_outbox (published_at);

-- +goose Down
DROP TABLE IF EXISTS orders_outbox;
