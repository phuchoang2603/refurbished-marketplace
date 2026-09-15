-- +goose Up
CREATE TABLE inventory_seed_intents (
    product_id UUID PRIMARY KEY,
    initial_qty INTEGER NOT NULL CHECK (initial_qty >= 0),
    product_version BIGINT NOT NULL CHECK (product_version = 1)
);

-- +goose Down
DROP TABLE inventory_seed_intents;
