-- +goose Up
TRUNCATE TABLE payment_outbox;

ALTER TABLE payment_outbox
  ALTER COLUMN payload TYPE BYTEA USING convert_to(payload::text, 'UTF8');

-- +goose Down
ALTER TABLE payment_outbox
  ALTER COLUMN payload TYPE JSONB USING convert_from(payload, 'UTF8')::jsonb;
