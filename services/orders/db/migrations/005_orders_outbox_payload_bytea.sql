-- +goose Up
-- Outbox payloads are canonical protobuf (orders service); JSONB rows cannot be converted reliably — clear before type change.
TRUNCATE TABLE orders_outbox;

ALTER TABLE orders_outbox
  ALTER COLUMN payload TYPE BYTEA USING convert_to(payload::text, 'UTF8');

-- +goose Down
ALTER TABLE orders_outbox
  ALTER COLUMN payload TYPE JSONB USING convert_from(payload, 'UTF8')::jsonb;
