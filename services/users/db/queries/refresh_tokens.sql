-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, token_hash, user_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING
    id, token_hash, user_id, expires_at, revoked_at, created_at, updated_at;

-- name: GetRefreshTokenByID :one
SELECT id, token_hash, user_id, expires_at, revoked_at, created_at, updated_at
FROM refresh_tokens
WHERE id = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1 AND revoked_at IS NULL
RETURNING
    id, token_hash, user_id, expires_at, revoked_at, created_at, updated_at;
