-- +goose Up
ALTER TABLE checkouts ADD COLUMN prep_operation_id UUID;
ALTER TABLE checkouts ADD COLUMN prep_version BIGINT;
ALTER TABLE checkouts ADD COLUMN prep_deadline_at TIMESTAMPTZ;
ALTER TABLE checkouts ADD COLUMN prep_closure TEXT CHECK (
    prep_closure IN ('PREPARED', 'REJECTED', 'EXPIRED')
);
ALTER TABLE checkouts ADD COLUMN prep_closed_at TIMESTAMPTZ;
ALTER TABLE checkouts ADD COLUMN orders_outcome TEXT NOT NULL DEFAULT 'PENDING' CHECK (
    orders_outcome IN ('PENDING', 'SUCCEEDED', 'REJECTED')
);
ALTER TABLE checkouts ADD COLUMN orders_reason TEXT;
ALTER TABLE checkouts ADD COLUMN inventory_outcome TEXT NOT NULL DEFAULT 'PENDING' CHECK (
    inventory_outcome IN ('PENDING', 'SUCCEEDED', 'REJECTED')
);
ALTER TABLE checkouts ADD COLUMN inventory_reason TEXT;

-- +goose Down
ALTER TABLE checkouts DROP COLUMN inventory_reason;
ALTER TABLE checkouts DROP COLUMN inventory_outcome;
ALTER TABLE checkouts DROP COLUMN orders_reason;
ALTER TABLE checkouts DROP COLUMN orders_outcome;
ALTER TABLE checkouts DROP COLUMN prep_closed_at;
ALTER TABLE checkouts DROP COLUMN prep_closure;
ALTER TABLE checkouts DROP COLUMN prep_deadline_at;
ALTER TABLE checkouts DROP COLUMN prep_version;
ALTER TABLE checkouts DROP COLUMN prep_operation_id;
