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

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, token_hash, user_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING refresh_tokens.*;

-- name: GetRefreshTokenByID :one
SELECT id, token_hash, user_id, expires_at, revoked_at, created_at, updated_at
FROM refresh_tokens
WHERE id = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1;
