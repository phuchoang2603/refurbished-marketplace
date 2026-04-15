# Refurbished Marketplace Plan

## Current Status

- Repository is a multi-module Go workspace (one `go.mod` per service, plus `shared/`).
- Initial services exist: `web`, `users`, `products`, `orders`, `cart`, `inventory`, **`payment`**.
- Architecture direction is explicit: REST at edge (`web` service), gRPC for internal service-to-service traffic.
- Development standardizes on Kubernetes (Tilt + Helm + CloudNativePG).
- Kafka dev stack (Strimzi + Kafka UI + Debezium connect) is deployed via Tilt/Helm; **domain consumers/producers for orders and payment are implemented**; full CDC/outbox→Kafka automation remains the long-term shape.
- **`shared/messaging`** exposes `NewKafkaConsumer` / `KafkaHandler` / `KafkaMessage` backed by **franz-go (`kgo`)** — pure Go, **no** `confluent-kafka-go` / librdkafka / CGO. Consumer uses `PollFetches`, **`Fetches.EachPartition`** + **`errgroup`** (parallel across partitions, ordered within a partition), **manual commits** after successful handler batches. **`BlockRebalanceOnPoll` is not used** (simpler operation; at-least-once delivery is assumed — **inbox / idempotent handlers** mitigate duplicates where it matters).
- **Bootstrap brokers**: `KafkaConsumerConfig.BootstrapServers []string`; env `KAFKA_BOOTSTRAP_SERVERS` is comma-split via **`messaging.ParseBootstrapServers`** in `cmd` binaries.
- **`users`** is the first implemented vertical slice (migration + sqlc + service + handlers + integration tests).
- Users auth: JWT login/refresh/logout and DB-backed refresh token sessions.
- **`web`** owns REST entrypoints; internal services are gRPC. **`web`** exposes **`POST /webhooks/stripe-simulator`** and forwards to **`payment`** over gRPC where applicable.
- **`orders`** is gRPC-first with PostgreSQL, sqlc, service tests, and per-item **`orders.item.created`** outbox rows keyed by `product_id`. It **consumes** **`payment.item.succeeded`** / **`payment.item.failed`** via Kafka and updates order status (today: **one event sets the whole-order status** — does not yet require all line items to succeed; refine later if product rules need “all items paid”).
- **`products`** is catalog-only; stock lives in **`inventory`**.
- **`inventory`** is stock/reservation (migrations + sqlc + service + tests).
- **`cart`** is ephemeral, backed by Redis/Valkey.
- **`payment`** is implemented: Postgres (intents, per-item transactions, **inbox**, **outbox**), gRPC API, Kafka consumer for **`orders.item.created`**, Stripe-simulator HTTP adapter, emits **`payment.item.*`** (outbox path for durable publish is in place; **Debezium wiring to Kafka** is still the intended production bridge).
  - Optional future: inventory-first gate event **`orders.pay_ready`** (key = `order_id`) so payment can be **reserve-then-pay** without aggregating item events.
- **Integration tests**: Kafka via Testcontainers **`confluentinc/confluent-local:7.5.0`** (aligned with **testcontainers-go**’s Kafka module). **REST/HTTP logs in that image are not where native Kafka produce/fetch shows up** — clients use the broker listener.
- **Dev shell**: **`flake.nix`** no longer pulls rdkafka/SASL/openssl for Go Kafka — franz-go is pure Go.

## Canonical Repo Tree

```text
/
  PLAN.md
  flake.nix
  Makefile
  Tiltfile
  services/
    web/
      cmd/web/
      internal/
        handlers/
        usersclient/
      tests/
      go.mod
    users/
    products/
    orders/
    cart/
    inventory/
    payment/
      cmd/payment/
      db/migrations/
      db/queries/
      internal/
      tests/
      go.mod
  shared/
    messaging/
    proto/
      users/v1/
      payment/v1/
      orders/v1/
      ...
    testutil/
    ...
  infra/
    charts/
      refurbished-marketplace/
      kafka/
    docker/
    k8s/
  docs/
```

## Stack and Conventions

- Language: Go, standard library first.
- Communication: REST only at edge/web layer, gRPC inside the microservice mesh.
- Database: PostgreSQL.
- Migrations: `goose`.
- Query generation: `sqlc`.
- Cache: Redis/Valkey (used by `cart`).
- Event bus: Kafka (Strimzi in dev); **Go clients use franz-go**, not the JVM REST stack inside `confluent-local`.
- Style: small packages, explicit SQL, straightforward handlers, table-driven tests.

## Service Layout Rules

- Start simple per service; avoid over-abstracting.
- Keep service code in `services/<name>/` and private code in `internal/`.
- Internal services expose gRPC contracts first and avoid new REST handlers.
- gRPC contracts live under `shared/proto/<domain>/v1/` and are generated into the same directories.
- Keep REST/HTTP DTO shaping in the web/edge service.
- Keep SQL and migrations service-local:
  - `services/<service>/db/migrations/`
  - `services/<service>/db/queries/`
- Keep all service tests in `services/<service>/tests/` (unit + integration).

## Focus Areas

- Keep `users` as identity/profile plus auth session source of truth.
- Keep `products` as catalog data with internal/admin writes.
- Keep `inventory` as the source of truth for available/reserved stock.
- Keep `orders` as order headers plus line items.
- Keep `cart` separate from order state and payment state.
- **`payment`** owns payment intents, per-item transactions, gateway simulation, inbox consumption of `orders.item.created`, and outbox for `payment.item.*`.

## Eventing Reliability

- Kafka is the async backbone for downstream consumers.
- `orders` writes one outbox row per item; `payment` uses a local **outbox** for **`payment.item.*`**.
- Debezium should stream outbox rows from Postgres into Kafka (operational wiring still the default target).
- **`payment`** uses an **inbox** for Kafka dedupe on `orders.item.created`.
- **`orders`** Kafka handler should stay safe under redelivery (ideally idempotent updates or future inbox if needed).
- Fraud/analytics should consume canonical streams, not ad hoc payloads.

## Minimal Schema

- `users`: `id`, `email`, `password_hash`
- `products`: `id`, `name`, `description`, `price_cents`, `merchant_id`
- `inventory`: `product_id`, `available_qty`, `reserved_qty`
- `orders`: `id`, `buyer_user_id`, `status`, `total_cents`
- `order_items`: `id`, `order_id`, `product_id`, `merchant_id`, `quantity`, `unit_price_cents`, `line_total_cents`
- `orders_outbox`: `id`, `aggregate_id`, `event_type`, `payload`, `publish_attempts`, `created_at`, `published_at`
- `cart`: Redis session state only; no Postgres schema required
- **`payment`**: intents, per-item transactions, inbox, outbox — see `services/payment/db/migrations/`

## Next Steps

1. Wire **Debezium / publisher** so `orders_outbox` and `payment_outbox` rows reliably reach Kafka topics (if not already complete in your cluster).
2. **Inventory**: consume `orders.item.created` keyed by `product_id` (if not done).
3. Aggregate **order payment status** from **all** line items when product rules require it (today: first `payment.item.succeeded` can mark the order paid).
4. Merchant snapshots / admin orchestration as needed.
5. Harden Kafka consumer options (**`BlockRebalanceOnPoll`**) only if you need stricter commit/rebalance coupling and accept the operational/test complexity.
