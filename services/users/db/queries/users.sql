-- name: CreateUser :one
INSERT INTO users (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING
    users.id,
    users.email,
    users.password_hash,
    users.created_at,
    users.updated_at;

-- name: GetUserByID :one
SELECT
    users.id,
    users.email,
    users.password_hash,
    users.created_at,
    users.updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT
    users.id,
    users.email,
    users.password_hash,
    users.created_at,
    users.updated_at
FROM users
WHERE email = $1;
