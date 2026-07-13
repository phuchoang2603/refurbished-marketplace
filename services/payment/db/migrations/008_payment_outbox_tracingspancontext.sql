-- +goose Up
ALTER TABLE payment_outbox
ADD COLUMN IF NOT EXISTS tracingspancontext TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE payment_outbox
DROP COLUMN IF EXISTS tracingspancontext;
