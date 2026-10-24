# Merchant-scoped checkout saga

Checkout is a durable, PostgreSQL-backed coordinator. Cart state is ephemeral; Orders owns orders, Inventory owns stock, and the external payment gateway owns hosted checkout and payment outcomes. No buyer-facing service directly reserves checkout stock or creates a payment session.

## Flow

```mermaid
sequenceDiagram
    autonumber
    participant B as Buyer
    participant W as Web
    participant C as Checkout
    participant K as Kafka (outbox)
    participant O as Orders
    participant I as Inventory
    participant G as Hosted gateway

    B->>W: Submit one merchant cart group (stable intent key)
    W->>W: Revalidate products, merchant, shipping, and totals
    W->>C: SubmitCheckout(snapshot, buyer, intent key)
    C->>C: Persist saga + outbox atomically
    C-->>W: Checkout ID
    W-->>B: Pending/progress page
    C->>K: checkout.prepare.requested.v1
    par Orders group
        K->>O: Create pending order or reject
        O->>K: checkout.order-created.v1 / checkout.order-rejected.v1
    and Inventory group
        K->>I: Reserve all lines or reject all
        I->>K: checkout.stock-reserved.v1 / checkout.stock-rejected.v1
    end
    K->>C: Correlated results, recorded independently
    C->>C: When both succeed, persist gateway operation with frozen snapshot
    C->>G: POST /v1/payments (Idempotency-Key = order_id)
    G-->>C: Payment ID, checkout URL, expiry
    C->>C: Record payment and READY_TO_PAY atomically
    B->>W: Poll buyer-scoped checkout progress
    W->>C: GetCheckout
    W-->>B: Redirect to gateway only when session ready
    B->>G: Complete payment
    G->>W: Signed POST /webhooks/payment-gateway (internal Service)
    W->>C: HandleGatewayWebhook(raw body, signature)
    C->>C: Verify HMAC and timestamp, deduplicate event ID
    C->>K: checkout.stock-commit.requested.v1 / checkout.stock-cancel.requested.v1
    K->>I: Commit or release hold
    I->>K: checkout.stock-committed.v1 / checkout.stock-released.v1
    K->>C: Settlement acknowledgement
    C->>K: checkout.order-finalize.requested.v1
    K->>O: Mark order paid / failed
    O->>K: checkout.order-finalized.v1
    K->>C: Finalization acknowledgement
    W->>W: Remove cart lines on SUCCEEDED; retain them and rotate the matching intent on EXPIRED
```

Orders and Inventory consume the same preparation command in distinct consumer groups, and Checkout joins their results durably. A rejection or the timeout worker claiming preparation after its two-minute recovery timeout closes preparation irreversibly. Both successes can still open payment after that timeout becomes due if they commit before compensation closes preparation. Once preparation is closed for compensation, Checkout cancels stock even if Inventory's outcome is unknown, waits for the released/absent acknowledgement, then finalizes the order as failed (Orders writes an absence fence if the order was never created) and waits for its acknowledgement. Late preparation successes never reopen payment. The buyer never sees a payment page in these cases. Gateway creation retries resend the stored request until the preparation deadline. `payment.succeeded` commits stock and `payment.expired` releases it; `attempt.declined` leaves payment open for another attempt. At gateway expiry plus grace, Checkout reads gateway status once. `SUCCEEDED` or `EXPIRED` settles stock; `OPEN` or any other non-terminal status or error leaves the workflow non-terminal with a `payment_uncertain` exception. Late contradictory success is recorded as a financial exception. There is no gateway cancellation API.

## Ownership and retries

- **Checkout** stores the saga, frozen gateway snapshot, durable gateway operations, inbox, outbox, deadlines, and exception records in PostgreSQL. It correlates each command/result to the checkout and retries durable due work after restarts.
- **Orders** creates orders from the shared preparation command and finalizes them only on Checkout commands; a failed finalization before creation leaves a durable fence that rejects late preparation. An order remains pending until Checkout receives stock settlement acknowledgement and issues finalization.
- **Inventory** reserves from the shared preparation command and accepts Checkout cancel/commit commands, fences cancel-before-reserve with an order tombstone, and returns durable idempotent results. `orders.created` does not reserve stock.
- **Gateway** hosts card entry and fraud decisions. Checkout alone holds its API key and webhook signing secret. Webhooks are HMAC-SHA256 signed with a ±5 minute timestamp tolerance and deduplicated by `event_id` in Checkout.
- **Web** submits a validated snapshot once per intent, displays buyer-scoped progress, forwards bounded raw gateway webhooks to Checkout, removes cart lines after observing `SUCCEEDED`, and retains them on `EXPIRED` while rotating only the matching checkout intent. The legacy direct `POST /orders`, `ReserveStock`, `CreateOrder`, `UpdateOrderStatus`, and `CreateHostedPaymentSession` RPCs are removed.

All command and result topics are versioned `checkout.*.v1`. Each producer's Debezium connector publishes its own outbox rows; consumers use inbox deduplication. The Checkout database is the source of truth for recovery, not a Kafka consumer offset or a browser redirect. See `services/checkout/README.md` for states and reconciliation rules.

## Rollout gate

Production cluster end-to-end verification is deferred. Before rollout, configure the gateway integration and matching marketplace credentials, then deploy with the new gateway configuration. The gateway owns its hosted hostname; the marketplace publishes only shop.

After rollout, verify approve, decline-then-approve, expiry, webhook redelivery and stock settlement using isolated fixtures. Confirm each event has one durable effect and accepted workflows remain buyer-queryable. Chart defaults have no gateway and return checkout unavailable. Local chart validation does not verify the live gateway integration or cluster workflows.

## Parallel preparation: scaling and cutover

Production runs three Orders and three Inventory replicas, each service with one stable `KAFKA_GROUP_ID` (set in the chart's service `env`, not a secret) shared by its pods, and six partitions on `checkout.prepare.requested.v1`. All other Kafka topics have one partition. Checkout-ID keys spread work across partitions; one checkout key cannot show parallelism. Inspect assignments with `kafka-consumer-groups.sh --describe --group <group>` for each service group (expect three members each with two prepare partitions apiece; consumers use the range balancer so each topic splits evenly). Match processing to pods using consumer spans, which carry `checkout_id`, `order_id`, Kafka partition/offset and duration, and the platform agent's pod attributes; submit checkouts spanning all partitions and confirm overlapping Orders and Inventory spans.

The initial create-order and reserve-stock commands are replaced, so cutover is not rolling-compatible:

1. Pause new checkout admission at the edge while keeping status reads, webhooks, consumers and workers running.
2. Drain accepted workflows through stock/order settlement; unresolved financial or compensating workflows block cutover. Confirm old command/result topic lag is zero and Debezium has published the outboxes.
3. Apply the migrations and provision `checkout.prepare.requested.v1`; the retired create-order and reserve-stock topics are removed from the chart in the same change.
4. Deploy Checkout, Orders, Inventory and Web together with admission still paused, verify both groups on the shared topic, then reopen admission.
