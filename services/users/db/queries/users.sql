-- name: CreateUser :one
INSERT INTO users (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING users.*;

-- name: GetUserByID :one
SELECT users.*
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT users.*
FROM users
WHERE email = $1;
