-- name: CreateProduct :one
INSERT INTO products (id, name, description, price_cents, stock, terminal_id, x_pos, y_pos)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING products.*;

-- name: GetProductByID :one
SELECT products.*
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT products.*
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
