# Database Schema

Each domain service owns its local database schema. The `products` service now owns catalog rows, stock rows, and reservation rows under one boundary. Cross-service references are logical IDs, not shared foreign-key ownership. Redis/Valkey is used for ephemeral cart state.

The current core order and payment model is merchant-scoped:

- cart items carry caller-supplied `merchant_id`
- each order belongs to exactly one merchant
- payment creates one transaction per order
- Kafka contracts use `orders.created`, `payment.succeeded`, and `payment.failed`
- product creation requires explicit initial stock
- reservation state is tracked inside the catalog boundary, not in a standalone inventory service

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

  INVENTORY {
    UUID product_id PK
    INTEGER available_qty
    INTEGER reserved_qty
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  INVENTORY_RESERVATIONS {
    UUID id PK
    UUID product_id FK
    UUID order_id
    INTEGER quantity
    TEXT status
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  ORDERS {
    UUID id PK
    UUID buyer_user_id
    UUID merchant_id
    TEXT status
    BIGINT total_cents
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  ORDER_ITEMS {
    UUID id PK
    UUID order_id FK
    UUID product_id
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

  USERS ||--o{ REFRESH_TOKENS : "owns"
  USERS ||--o{ ORDERS : "buys"
  ORDERS ||--o{ ORDER_ITEMS : "contains"
  ORDERS ||--|| PAYMENT_INTENTS : "settles"
  PAYMENT_INTENTS ||--o{ PAYMENT_TRANSACTIONS : "spawns"
  PRODUCTS ||--o{ ORDER_ITEMS : "referenced by"
  PRODUCTS ||--|| INVENTORY : "tracks"
  PRODUCTS ||--o{ INVENTORY_RESERVATIONS : "reserves"
  ORDERS ||--o{ INVENTORY_RESERVATIONS : "creates"
```
