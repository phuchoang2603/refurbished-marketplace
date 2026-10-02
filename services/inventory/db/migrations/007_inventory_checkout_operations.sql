-- +goose Up
CREATE TABLE inventory_checkout_operations (
    order_id UUID PRIMARY KEY,
    checkout_id UUID NOT NULL UNIQUE,
    reserve_hash BYTEA,
    status TEXT NOT NULL CHECK (
        status IN ('RESERVED', 'REJECTED', 'ABSENT', 'RELEASED', 'COMMITTED')
    ),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE inventory_checkout_operations;
