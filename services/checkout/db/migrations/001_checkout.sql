-- +goose Up
CREATE TABLE checkouts (
    id UUID PRIMARY KEY,
    buyer_user_id UUID NOT NULL,
    merchant_id UUID NOT NULL,
    intent_key TEXT NOT NULL,
    request_hash BYTEA NOT NULL,
    snapshot JSONB NOT NULL,
    state TEXT NOT NULL,
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    order_id UUID UNIQUE,
    payment_session_id TEXT,
    payment_return_url TEXT,
    failure_reason TEXT,
    deadline_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (buyer_user_id, intent_key),
    CHECK (JSONB_TYPEOF(snapshot) = 'object')
);

CREATE INDEX checkouts_due_idx ON checkouts (
    deadline_at, id
) WHERE deadline_at IS NOT NULL;
CREATE INDEX checkouts_pending_age_idx ON checkouts (state, updated_at);

CREATE TABLE checkout_inbox (
    message_id TEXT PRIMARY KEY,
    checkout_id UUID NOT NULL REFERENCES checkouts(id),
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX checkout_inbox_checkout_id_idx ON checkout_inbox (checkout_id);

CREATE TABLE checkout_outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL REFERENCES checkouts(id),
    event_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    tracingspancontext TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX checkout_outbox_aggregate_id_idx ON checkout_outbox (aggregate_id);

CREATE TABLE checkout_exceptions (
    id UUID PRIMARY KEY,
    checkout_id UUID NOT NULL REFERENCES checkouts(id),
    message_id TEXT,
    kind TEXT NOT NULL,
    details TEXT NOT NULL,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (checkout_id, message_id, kind)
);
CREATE INDEX checkout_exceptions_unresolved_idx ON checkout_exceptions (
    created_at
) WHERE resolved_at IS NULL;

-- +goose Down
DROP TABLE checkout_exceptions;
DROP TABLE checkout_outbox;
DROP TABLE checkout_inbox;
DROP TABLE checkouts;
