-- +goose Up
CREATE TABLE IF NOT EXISTS inventory_reservations (
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (order_id, product_id),
    CONSTRAINT inventory_reservations_quantity_positive CHECK (quantity > 0),
    CONSTRAINT inventory_reservations_status_valid CHECK (
        status IN ('RESERVED', 'COMMITTED', 'RELEASED')
    ),
    CONSTRAINT inventory_reservations_product_id_fk FOREIGN KEY (
        product_id
    ) REFERENCES products (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS inventory_reservations_status_idx ON inventory_reservations (
    status
);

-- +goose Down
DROP INDEX IF EXISTS inventory_reservations_status_idx;
DROP TABLE IF EXISTS inventory_reservations;
