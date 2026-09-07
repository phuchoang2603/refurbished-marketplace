-- +goose Up
UPDATE payment_intents
SET status = 'FAILED'
WHERE status = 'CANCELLED';

ALTER TABLE payment_intents
DROP COLUMN IF EXISTS cancel_url;

-- +goose Down
ALTER TABLE payment_intents
ADD COLUMN IF NOT EXISTS cancel_url TEXT NOT NULL DEFAULT '';
