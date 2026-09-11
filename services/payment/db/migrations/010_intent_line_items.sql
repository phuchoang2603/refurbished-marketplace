-- +goose Up
ALTER TABLE payment_intents
ADD COLUMN IF NOT EXISTS line_items JSONB NOT NULL DEFAULT '[]'::JSONB;

-- +goose Down
ALTER TABLE payment_intents DROP COLUMN IF EXISTS line_items;
