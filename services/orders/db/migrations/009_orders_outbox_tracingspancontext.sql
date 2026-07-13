-- +goose Up
ALTER TABLE orders_outbox
ADD COLUMN IF NOT EXISTS tracingspancontext TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE orders_outbox
DROP COLUMN IF EXISTS tracingspancontext;
