-- name: CreateProduct :one
INSERT INTO products (id, name, description, price_cents, stock)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, price_cents, stock, created_at, updated_at;

-- name: GetProductByID :one
SELECT id, name, description, price_cents, stock, created_at, updated_at
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT id, name, description, price_cents, stock, created_at, updated_at
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
