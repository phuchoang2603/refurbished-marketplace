-- name: CreateProduct :one
INSERT INTO products (
    id,
    name,
    description,
    price_cents,
    merchant_id
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    id, name, description, price_cents, created_at, updated_at, merchant_id;

-- name: GetProductByID :one
SELECT id, name, description, price_cents, created_at, updated_at, merchant_id
FROM
    products
WHERE
    id = $1;

-- name: ListProducts :many
SELECT
    id, name, description, price_cents, created_at, updated_at, merchant_id
FROM
    products
ORDER BY
    created_at DESC, id DESC
LIMIT $1 OFFSET $2;
