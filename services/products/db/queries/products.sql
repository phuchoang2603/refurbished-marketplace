-- name: CreateProduct :one
INSERT INTO products (id, owner_user_id, name, description, price_cents, stock)
VALUES ($1, $2, $3, $4, $5, $6)
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

-- name: UpdateProductByIDAndOwner :one
UPDATE products
SET
  name = COALESCE(sqlc.narg('name')::text, name),
  description = COALESCE(sqlc.narg('description')::text, description),
  price_cents = COALESCE(sqlc.narg('price_cents')::bigint, price_cents),
  stock = COALESCE(sqlc.narg('stock')::integer, stock),
  updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND owner_user_id = sqlc.arg('owner_user_id')
RETURNING products.*;

-- name: DeleteProductByIDAndOwner :execrows
DELETE FROM products
WHERE id = $1
  AND owner_user_id = $2;
