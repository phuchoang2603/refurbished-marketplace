-- +goose Up
ALTER TABLE payment_intents
ADD COLUMN IF NOT EXISTS payment_session_id TEXT,
ADD COLUMN IF NOT EXISTS return_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS cancel_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS failure_reason TEXT;

CREATE INDEX IF NOT EXISTS payment_intents_payment_session_id_idx ON payment_intents (
    payment_session_id
);

-- +goose Down
DROP INDEX IF EXISTS payment_intents_payment_session_id_idx;

ALTER TABLE payment_intents
DROP COLUMN IF EXISTS failure_reason,
DROP COLUMN IF EXISTS expires_at,
DROP COLUMN IF EXISTS cancel_url,
DROP COLUMN IF EXISTS return_url,
DROP COLUMN IF EXISTS payment_session_id;
