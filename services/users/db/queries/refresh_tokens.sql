-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, token_hash, user_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING refresh_tokens.*;

-- name: GetRefreshTokenByID :one
SELECT *
FROM refresh_tokens
WHERE id = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1 AND revoked_at IS NULL
RETURNING refresh_tokens.*;
