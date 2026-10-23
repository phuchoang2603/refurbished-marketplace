-- +goose Up
ALTER TABLE checkouts RENAME COLUMN payment_session_id TO gateway_payment_id;
ALTER TABLE checkouts RENAME COLUMN payment_return_url TO hosted_checkout_url;
ALTER TABLE checkouts ADD COLUMN gateway_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB;
ALTER TABLE checkouts ADD COLUMN gateway_expires_at TIMESTAMPTZ;
ALTER TABLE checkouts ADD COLUMN gateway_last_event_at TIMESTAMPTZ;
ALTER TABLE checkouts ADD COLUMN gateway_status_checked_at TIMESTAMPTZ;
ALTER TABLE checkouts ADD COLUMN payment_status TEXT;
CREATE UNIQUE INDEX checkouts_gateway_payment_idx ON checkouts (
    gateway_payment_id
) WHERE gateway_payment_id IS NOT NULL;
CREATE TABLE checkout_gateway_operations (
    checkout_id UUID PRIMARY KEY REFERENCES checkouts(id),
    request BYTEA NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'succeeded', 'failed')
    ),
    tracingspancontext TEXT
);
CREATE INDEX checkout_gateway_operations_due_idx ON checkout_gateway_operations (
    next_attempt_at
) WHERE status
= 'pending';

-- +goose Down
DROP TABLE checkout_gateway_operations;
DROP INDEX checkouts_gateway_payment_idx;
ALTER TABLE checkouts DROP COLUMN gateway_snapshot;
ALTER TABLE checkouts DROP COLUMN gateway_expires_at;
ALTER TABLE checkouts DROP COLUMN gateway_last_event_at;
ALTER TABLE checkouts DROP COLUMN gateway_status_checked_at;
ALTER TABLE checkouts DROP COLUMN payment_status;
ALTER TABLE checkouts RENAME COLUMN gateway_payment_id TO payment_session_id;
ALTER TABLE checkouts RENAME COLUMN hosted_checkout_url TO payment_return_url;
