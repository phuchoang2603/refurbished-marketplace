# Users Auth Implementation Plan

## Goal

Add JWT-based login and refresh flows to the `users` service using one shared JWT secret for simplicity, while storing refresh token sessions in PostgreSQL for revocation and rotation.

## Scope

- Service: `services/users`
- Transport: HTTP REST (existing users handlers)
- Persistence: PostgreSQL (`goose` + `sqlc`)
- Token strategy:
  - Access token: short-lived JWT
  - Refresh token: long-lived JWT with DB-backed session state

## Non-Goals (for this phase)

- OAuth/social login
- RBAC/permission system
- Moving auth logic to `shared/`
- Istio policy integration details

## Configuration

Use one secret env var:

- `JWT_SECRET` (required)

Keep these as code defaults (no required env vars):

- issuer: `refurbished-marketplace`
- audience: `refurbished-marketplace-api`
- access TTL: `15m`
- refresh TTL: `168h`

## Data Model

Add `refresh_tokens` table via Goose migration.

Recommended schema shape:

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `token_hash TEXT NOT NULL UNIQUE`
- `expires_at TIMESTAMPTZ NOT NULL`
- `revoked_at TIMESTAMPTZ`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Indexes:

- `refresh_tokens_user_id_idx(user_id)`
- `refresh_tokens_expires_at_idx(expires_at)`

Why hash refresh tokens:

- If DB is leaked, raw refresh tokens are not immediately usable.

## API Endpoints

Add endpoints under `users` service:

- `POST /auth/login`
  - Input: `email`, `password`
  - Output: `access_token`, `refresh_token`, `token_type`, `expires_in`
- `POST /auth/refresh`
  - Input: `refresh_token`
  - Output: new `access_token`, new `refresh_token`, `token_type`, `expires_in`
- `POST /auth/logout` (optional in this phase)
  - Input: `refresh_token`
  - Output: `204`

## Token and Session Flow

### Login

1. Get user by email.
2. Verify password with bcrypt.
3. Create refresh session row with `token_hash` and expiry.
4. Issue access and refresh JWTs.

### Refresh

1. Verify refresh JWT signature/claims.
2. Validate refresh session exists, not revoked, not expired.
3. Rotate session:
   - revoke old session
   - create new session
4. Issue new access and refresh tokens.

### Logout

1. Verify refresh token.
2. Revoke matching session.

## JWT Claims

Minimal claim set:

- `sub`: user ID
- `aud`: `refurbished-marketplace-api`
- `iss`: `refurbished-marketplace`
- `exp`, `iat`
- `jti`: session token id (for refresh token tracking)
- `typ`: `access` or `refresh`

## Implementation Breakdown

1. Migration: add `refresh_tokens` table.
2. SQL (`sqlc`): create/get/revoke refresh session queries.
3. Service layer:
   - `Login(email, password)`
   - `Refresh(refreshToken)`
   - `Logout(refreshToken)`
4. JWT helper package (inside users service): sign/verify tokens.
5. HTTP handlers for auth endpoints.
6. Tests in `services/users/tests/` only:
   - service behavior tests
   - integration tests with Testcontainers + Goose

## Error Mapping

- Invalid credentials -> `401`
- Invalid/expired refresh token -> `401`
- Revoked session -> `401`
- Validation errors -> `400`
- Internal errors -> `500`

## Istio Consideration

Istio may validate access JWTs later, but this plan keeps issuance/refresh/revocation in the users service where it belongs.
