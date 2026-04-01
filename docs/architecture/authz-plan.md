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

- Owns catalog resource rules:
  - public reads for browsing
  - internal-only write paths if catalog management is reintroduced later

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

## Data and Contract Changes

Products model should focus on catalog and terminal metadata, not public ownership writes.

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
