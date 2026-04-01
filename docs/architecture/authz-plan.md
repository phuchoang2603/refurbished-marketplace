# Web Authorization Plan

## Goal

Use `web` as the authorization gateway for REST endpoints while keeping auth session logic in `users` service.

This repository is scoped as a normal ecommerce platform, not C2C escrow.

## Responsibility Split

### users service

- Owns identity/session domain logic:
  - `Login`
  - `Refresh`
  - `Logout`
- Issues JWT tokens and manages refresh-token persistence/revocation.

### web service

- Owns REST authorization decisions:
  - validate bearer access JWT on protected routes
  - extract `sub` user id claim
  - propagate trusted user id to internal gRPC calls

### products service

- Owns resource authorization invariants:
  - persist `owner_user_id` for each product
  - enforce owner-only write semantics for update/delete

## Payments / Events

- Orders will emit domain events for downstream payment/fraud/recommender consumers.
- Prefer Kafka/Strimzi for async consumers and ML integrations.
- CDC is not the default starting point; prefer explicit domain events first.

## Scope Note

- Recommender is a later consumer of product/order events.
- External payment/fraud platform is a separate project and should be integrated by API/event contracts later.

## Endpoint Access Policy

Public:

- `GET /healthz`
- `POST /auth/login`
- `POST /auth/refresh`
- `GET /products`
- `GET /products/{id}`

Authenticated:

- `POST /auth/logout`
- `POST /products`
- `PATCH /products/{id}`
- `DELETE /products/{id}`

## Data and Contract Changes

Products model should include:

- `owner_user_id UUID NOT NULL`

Products gRPC should include:

- `owner_user_id` in `Product`
- `owner_user_id` in mutation requests (trusted input from web)
- update/delete methods for owner mutations

## Why Not Move Everything to Web

- Avoid duplicating session/token lifecycle logic.
- Keep identity domain in one place (`users`).
- Preserve cleaner service boundaries for future entrypoints.

## Shared Package Decision

Use a small shared JWT verification package only for common cryptographic validation and claim parsing.

Recommended constraints:

- Keep it stateless and narrow (parse/verify claims only).
- Do not move users session DB logic or refresh orchestration to shared.
- Keep web-specific authorization middleware in `services/web/internal`.

Suggested location:

- `shared/auth/jwt` for claims + token verify helpers.

This avoids code drift between users/web while preserving clear domain ownership.
