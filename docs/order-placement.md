# Merchant-Scoped Order, Inventory, And Payment Flow

This document describes the current core contract between `cart`, `orders`, `inventory`, and `payment`, including the hosted-payment redirect flow used in development.

## Core Model

- `cart` stores ephemeral cart state and requires web-supplied `merchant_id` on item writes.
- `orders` accepts only merchant-scoped order creation requests.
- `inventory` reserves stock per order before payment proceeds.
- `payment` creates or reuses one hosted payment session per order and creates one payment transaction per order after reservation.

## Flow

```mermaid
sequenceDiagram
    autonumber
    participant WEB as Web
    participant SIM as Simulator
    participant CRT as Cart
    participant ORD as Orders
    participant K as Kafka
    participant INV as Inventory
    participant PAY as Payment

    Note over WEB, CRT: Ephemeral Phase
    WEB->>CRT: AddCartItem(merchant_id, qty)
    CRT-->>WEB: 200 OK

    Note over WEB, ORD: Order Creation
    WEB->>ORD: CreateOrder(merchant_id, total)
    ORD->>ORD: Persist Order (Status: Pending)
    ORD->>K: Emit "orders.created"
    ORD-->>WEB: 201 Created

    Note over WEB, SIM: Hosted Payment Setup
    WEB->>PAY: CreateHostedPaymentSession(order_id, buyer, shipping, return URLs)
    PAY->>PAY: Persist payment session(order_id)
    PAY-->>WEB: payment session metadata
    WEB->>WEB: Build hosted payment URL
    WEB->>SIM: Redirect browser to hosted payment URL
    SIM->>WEB: POST hosted payment callback
    WEB->>PAY: HandleGatewayWebhook
    SIM->>WEB: Redirect browser back to /orders/{id}

    Note over K, INV: Reservation Loop
    K->>INV: Consume "orders.created"
    INV->>INV: Record inventory_inbox message
    INV->>INV: Reserve order item stock
    INV->>K: Emit "inventory.reserved"

    Note over K, PAY: Payment Loop
    K->>PAY: Consume "inventory.reserved"
    PAY->>PAY: Record payment_inbox message
    PAY->>PAY: Load payment_intent(order_id)
    PAY->>PAY: Create payment_transaction(order_id)
    PAY->>K: Emit "payment.succeeded"
    K->>INV: Consume "payment.succeeded"
    INV->>INV: Commit reservation
    K->>ORD: Consume "payment.succeeded"
    ORD->>ORD: Update Status: Paid

    Note over K, ORD: Reservation Failure Loop
    INV->>K: Emit "inventory.reservation-failed"
    K->>ORD: Consume "inventory.reservation-failed"
    ORD->>ORD: Update Status: Failed
```

## Responsibilities

### Web

- Orchestrates browser checkout, order creation, and hosted payment redirect.
- Builds the buyer-facing hosted payment URL from payment session metadata and gateway configuration.
- Accepts hosted gateway callbacks and forwards terminal outcomes to `payment` over gRPC.

### Cart

- Stores `product_id`, `merchant_id`, and `quantity` in Redis/Valkey-backed state.
- Validates that `cart_id`, `product_id`, and `merchant_id` are present and UUID-shaped.
- Does not derive merchant ownership from `products`.

### Orders

- Persists one order per merchant.
- Stores `merchant_id` on the order record.
- Stores order items with `product_id`, `quantity`, `unit_price_cents`, and `line_total_cents`.
- Emits one `orders.created` outbox event per created order, including item lines.
- Consumes `inventory.reservation-failed`, `payment.succeeded`, and `payment.failed` to update order status.

### Inventory

- Stores aggregate stock in `inventory` and reservation ownership in inventory-local reservation records.
- Consumes `orders.created` and reserves all order item lines idempotently per `order_id`.
- Emits `inventory.reserved` when the order is fully reserved.
- Emits `inventory.reservation-failed` when the order cannot be fully reserved.
- Consumes `payment.succeeded` and `payment.failed` to commit or release reserved stock.

### Payment

- Stores hosted payment session state by `order_id`.
- Reuses `order_id` as the idempotency anchor for hosted session creation.
- Consumes `inventory.reserved`.
- Creates one payment transaction per `order_id`.
- Applies hosted gateway outcomes from the web edge and emits `payment.succeeded` or `payment.failed` through the payment outbox.

### Dev Simulator

- A dev-only hosted payment simulator can live under `tools/`.
- It renders a mock hosted payment page, posts a terminal callback to `services/web`, and redirects the browser back to the marketplace order page.

## Event Contracts

### `orders.created`

Produced by `orders` once per created order.

Carries:

- `order_id`
- `buyer_user_id`
- `merchant_id`
- `total_cents`
- `items[]` with `product_id` and `quantity`

### `inventory.reserved`

Produced by `inventory` once per fully reserved order.

Carries:

- `order_id`
- `merchant_id`
- `total_cents`

### `inventory.reservation-failed`

Produced by `inventory` once per order that cannot be fully reserved.

Carries:

- `order_id`

### `payment.succeeded` / `payment.failed`

Produced by `payment` once per payment transaction outcome.

Carries:

- `order_id`
