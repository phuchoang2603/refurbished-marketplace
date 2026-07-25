-- +goose Up
ALTER TABLE inventory_outbox
ADD COLUMN IF NOT EXISTS tracingspancontext TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE inventory_outbox
DROP COLUMN IF EXISTS tracingspancontext;
