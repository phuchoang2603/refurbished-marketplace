# Repository Architecture

This repository is a multi-module Go workspace for a multi-service marketplace.

## Boundaries

- `web` owns the public REST edge.
- Internal services use gRPC.
- Kafka handles async integration between domain services.
- PostgreSQL owns service-local persistence.
- Redis/Valkey powers the ephemeral cart.

## Runtime Layout

- Go services are kept under `services/<name>/`.
- Shared protobuf contracts live under `shared/proto/<domain>/v1/`.
- Local development is managed with `devenv.nix`, plus Tilt and Helm for the Kubernetes stack.

## System Flow

```mermaid
graph LR
  Client[Client] --> Web[web]
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

## Database ERD

The diagram below reflects the tables created by each service migration. Some cross-service links are logical IDs rather than enforced foreign keys because each service owns its own database.

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

  ORDERS_OUTBOX {
    UUID id PK
    UUID aggregate_id
    TEXT event_type
    BYTEA payload
    INTEGER publish_attempts
    TIMESTAMPTZ created_at
    TIMESTAMPTZ published_at
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

  PAYMENT_INBOX {
    TEXT message_id PK
    TIMESTAMPTZ received_at
  }

  PAYMENT_OUTBOX {
    UUID id PK
    UUID aggregate_id
    TEXT event_type
    BYTEA payload
    INTEGER publish_attempts
    TIMESTAMPTZ created_at
    TIMESTAMPTZ published_at
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
