# Refurbished Marketplace

## Overview

The system models a marketplace for browsing products, managing carts, creating orders, coordinating inventory, and handling payment flows. Internal services communicate over gRPC. Domain events and eventual consistency workflows use Kafka, while each service owns its own persistence boundary.

## Architecture

### Service Boundaries

- `services/web` owns the public browser edge, auth boundary, server-rendered pages, and Datastar-compatible fragments.
- `services/users` owns users, credentials, access tokens, and refresh-token sessions.
- `services/products` owns product catalog data.
- `services/cart` owns ephemeral cart state.
- `services/orders` owns order creation, order state, and order outbox events.
- `services/inventory` owns inventory availability and reservation state.
- `services/payment` owns payment intents, payment transactions, and gateway webhook handling.
- `shared/proto` contains the protobuf contracts used by service clients and servers.

### System Flow

```mermaid
graph LR
  Browser[Browser] --> Web[web]
  Web --> Users[users]
  Web --> Products[products]
  Web --> Cart[cart]
  Web --> Orders[orders]
  Orders --> Kafka[(Kafka)]
  Kafka --> Inventory[inventory]
  Kafka --> Payment[payment]
  Kafka --> Orders
  Payment --> Kafka
  Inventory --> Kafka
```

### Data Ownership

Each domain service owns its local database schema. Cross-service references are logical IDs, not shared foreign-key ownership. Redis/Valkey is used for ephemeral cart state.

```mermaid
erDiagram
  USERS {
    UUID id PK
    TEXT email
    TEXT password_hash
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  REFRESH_TOKENS {
    UUID id PK
    TEXT token_hash
    UUID user_id FK
    TIMESTAMPTZ expires_at
    TIMESTAMPTZ revoked_at
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  PRODUCTS {
    UUID id PK
    TEXT name
    TEXT description
    BIGINT price_cents
    UUID merchant_id
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  ORDERS {
    UUID id PK
    UUID buyer_user_id
    TEXT status
    BIGINT total_cents
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  ORDER_ITEMS {
    UUID id PK
    UUID order_id FK
    UUID product_id
    UUID merchant_id
    INTEGER quantity
    BIGINT unit_price_cents
    BIGINT line_total_cents
    TIMESTAMPTZ created_at
  }

  PAYMENT_INTENTS {
    UUID order_id PK
    UUID buyer_user_id
    TEXT payment_token
    TEXT currency
    JSONB billing_address
    JSONB shipping_address
    TEXT status
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  PAYMENT_TRANSACTIONS {
    UUID id PK
    UUID order_id FK
    UUID order_item_id
    UUID merchant_id
    BIGINT amount_cents
    TEXT currency
    TEXT status
    TEXT idempotency_key
    TEXT gateway_transaction_id
    TEXT failure_reason
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  INVENTORY {
    UUID product_id PK
    INTEGER available_qty
    INTEGER reserved_qty
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  USERS ||--o{ REFRESH_TOKENS : "owns"
  USERS ||--o{ ORDERS : "buys"
  ORDERS ||--o{ ORDER_ITEMS : "contains"
  ORDERS ||--|| PAYMENT_INTENTS : "settles"
  PAYMENT_INTENTS ||--o{ PAYMENT_TRANSACTIONS : "spawns"
  PRODUCTS ||--o{ ORDER_ITEMS : "referenced by"
  PRODUCTS ||--|| INVENTORY : "tracks"
```

## Tech Stack

- Go for all services and shared libraries.
- gRPC and Protocol Buffers for internal service APIs.
- PostgreSQL for service-local durable persistence, `sqlc` for queries generation and `goose` for migration
- Redis/Valkey for cart state.
- Kafka for asynchronous domain integration.
- `templ` for typed server-rendered HTML components.
- Datastar-compatible markup for browser interactions and fragment updates.
- Tilt, Helm, and Kubernetes manifests for local/runtime orchestration.
- Nix/devenv for local development environment setup.
- OpenSpec for change proposals, specs, designs, tasks, and archives.

## Development

This repository uses `devenv` to install and pin local tooling. Enter the development shell before running generators, tests, or local infrastructure commands:

```bash
devenv shell
```

The shell provides the project tooling defined in `devenv.nix`, such as Go, protobuf tooling, database migration/query generators, Kubernetes tooling, and related CLIs. Prefer adding new required developer tools to `devenv.nix` instead of relying on globally installed binaries.

Local Kubernetes development is managed with Tilt. After entering the `devenv` shell, start the stack with:

```bash
tilt up
```

Tilt uses the root `Tiltfile` to build services, apply the Kubernetes/Helm resources under `infra/`, and keep the local cluster in sync while you edit code. Use the Tilt UI to inspect service status, logs, resource readiness, and rebuilds.

### Repository Layout

- `services/<name>/`: service implementations.
- `shared/`: shared Go modules and protobuf contracts.
- `infra/`: deployment and infrastructure configuration.
- `openspec/`: active and archived change artifacts.
- `docs/`: project documentation.
