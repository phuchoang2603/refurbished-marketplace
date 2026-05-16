# Merchant-Scoped Order, Inventory, And Payment Flow

This document describes the current core contract between `cart`, `orders`, `inventory`, and `payment`.

## Core Model

- `cart` stores ephemeral cart state and requires caller-supplied `merchant_id` on item writes.
- `orders` accepts only merchant-scoped order creation requests.
- `inventory` reserves stock per order before payment proceeds.
- `payment` creates one payment transaction per order.

## Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Caller
    participant CRT as Cart
    participant ORD as Orders
    participant K as Kafka
    participant INV as Inventory
    participant PAY as Payment

    Note over C, CRT: Ephemeral Phase
    C->>CRT: AddCartItem(merchant_id, qty)
    CRT-->>C: 200 OK

    Note over C, PAY: Payment Setup
    C->>PAY: InitiatePayment(order_id, buyer, token, addresses)
    PAY->>PAY: Persist payment_intent(order_id)
    PAY-->>C: 200 OK

    Note over C, ORD: Order Creation
    C->>ORD: CreateOrder(merchant_id, total)
    ORD->>ORD: Persist Order (Status: Pending)
    ORD->>K: Emit "orders.created"
    ORD-->>C: 201 Created

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
    INV->>K: Emit "inventory.reservation_failed"
    K->>ORD: Consume "inventory.reservation_failed"
    ORD->>ORD: Update Status: Failed
```

## Responsibilities

### Cart

- Stores `product_id`, `merchant_id`, and `quantity` in Redis/Valkey-backed state.
- Validates that `cart_id`, `product_id`, and `merchant_id` are present and UUID-shaped.
- Does not derive merchant ownership from `products`.

### Orders

- Persists one order per merchant.
- Stores `merchant_id` on the order record.
- Stores order items with `product_id`, `quantity`, `unit_price_cents`, and `line_total_cents`.
- Emits one `orders.created` outbox event per created order, including item lines.
- Consumes `inventory.reservation_failed`, `payment.succeeded`, and `payment.failed` to update order status.

### Inventory

- Stores aggregate stock in `inventory` and reservation ownership in inventory-local reservation records.
- Consumes `orders.created` and reserves all order item lines idempotently per `order_id`.
- Emits `inventory.reserved` when the order is fully reserved.
- Emits `inventory.reservation_failed` when the order cannot be fully reserved.
- Consumes `payment.succeeded` and `payment.failed` to commit or release reserved stock.

### Payment

- Stores payment intent by `order_id`.
- Consumes `inventory.reserved`.
- Creates one payment transaction per `order_id`.
- Emits `payment.succeeded` or `payment.failed` through the payment outbox.

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

### `inventory.reservation_failed`

Produced by `inventory` once per order that cannot be fully reserved.

Carries:

- `order_id`

### `payment.succeeded` / `payment.failed`

Produced by `payment` once per payment transaction outcome.

Carries:

- `order_id`
