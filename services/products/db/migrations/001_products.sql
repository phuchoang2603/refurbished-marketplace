-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL CHECK (
        category IN (
            'phone',
            'laptop',
            'tablet',
            'accessory',
            'headphones',
            'smartwatch',
            'game_console',
            'camera'
        )
    ),
    condition TEXT NOT NULL CHECK (condition IN ('excellent', 'good', 'fair')),
    price_cents BIGINT NOT NULL CHECK (price_cents > 0),
    merchant_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX products_merchant_id_idx ON products (merchant_id);

CREATE TABLE products_outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL REFERENCES products(id),
    event_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    tracingspancontext TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE products_outbox;
DROP TABLE products;
