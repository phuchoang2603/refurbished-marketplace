-- +goose Up
CREATE TABLE payment_checkout_sessions (
    order_id UUID PRIMARY KEY,
    checkout_id UUID NOT NULL UNIQUE,
    request_hash BYTEA,
    create_operation_id UUID,
    create_version BIGINT,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'ABSENT')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_checkout_callbacks (
    order_id UUID NOT NULL REFERENCES payment_checkout_sessions(order_id),
    outcome TEXT NOT NULL CHECK (outcome IN ('SUCCEEDED', 'FAILED')),
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (order_id, outcome)
);

-- +goose Down
DROP TABLE payment_checkout_callbacks;
DROP TABLE payment_checkout_sessions;
