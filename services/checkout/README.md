# Checkout workflow

Checkout persists the buyer's accepted, merchant-scoped commerce snapshot and is the only coordinator for new checkouts. Orders owns the order, Inventory owns stock and reservations, and the external payment gateway owns the hosted payment and attempt outcomes. A browser disconnect does not cancel an accepted workflow.

| Checkout state                                 | Allowed correlated result      | Durable command or transition                                              | Owner                |
| ---------------------------------------------- | ------------------------------ | -------------------------------------------------------------------------- | -------------------- |
| CREATING_ORDER                                 | OrderCreated                   | Request reservation; RESERVING_STOCK                                       | Checkout → Inventory |
| CREATING_ORDER                                 | OrderRejected                  | Record definitive failure; FAILED                                          | Checkout             |
| CREATING_ORDER                                 | Deadline, result unknown       | Retry the same order operation                                             | Checkout → Orders    |
| RESERVING_STOCK                                | StockReserved                  | Persist gateway operation; CREATING_PAYMENT                                | Checkout             |
| RESERVING_STOCK                                | StockRejected or deadline      | Fence reservation; RELEASING_STOCK                                         | Checkout → Inventory |
| CREATING_PAYMENT                               | Gateway creation succeeds      | Persist payment ID, hosted URL and expiry; READY_TO_PAY                    | Checkout             |
| CREATING_PAYMENT                               | Defect or preparation deadline | Release stock; RELEASING_STOCK                                             | Checkout → Inventory |
| READY_TO_PAY                                   | attempt.declined               | Record event time; keep payment open                                       | Checkout             |
| READY_TO_PAY or NEEDS_REVIEW with OPEN payment | payment.succeeded              | Commit stock; COMMITTING_STOCK                                             | Checkout → Inventory |
| READY_TO_PAY or NEEDS_REVIEW with OPEN payment | payment.expired                | Release stock; RELEASING_STOCK                                             | Checkout → Inventory |
| READY_TO_PAY                                   | Expiry plus grace              | Read gateway status once; terminal outcome settles, otherwise NEEDS_REVIEW | Checkout → gateway   |
| COMMITTING_STOCK                               | StockCommitted                 | Finalize order paid; FINALIZING_PAID                                       | Checkout → Orders    |
| RELEASING_STOCK                                | StockReleased or absent        | Finalize order failed; FINALIZING_FAILED                                   | Checkout → Orders    |
| FINALIZING_PAID                                | OrderFinalized(PAID)           | COMPLETED                                                                  | Checkout             |
| FINALIZING_FAILED                              | OrderFinalized(FAILED)         | FAILED                                                                     | Checkout             |
| Any                                            | Duplicate result/event         | Acknowledge without repeating effects                                      | Checkout             |
| Settling or terminal                           | Contradictory verified payment | Record financial exception; pause nonterminal workflow for review          | Checkout             |

Kafka results carry checkout, order, message and operation IDs and a workflow version. Stale results cannot roll back a workflow. Stock cancellation persists a tombstone even if it precedes reserve. Inbox, state and outbox changes commit together.

At submission, Checkout reads seller and buyer account creation times in one Redis lookup from its local account projection and freezes the authoritative product grading, prices, address and HTTPS return URL. Buyer email is omitted. The stock-reserved transaction stores the gateway request bytes in `checkout_gateway_operations`. Workers lease due operations with `SKIP LOCKED`, use the order ID as `Idempotency-Key`, and resend identical bytes after transport errors or lost local commits. Creation defects and preparation deadlines compensate; no hosted URL is exposed until readiness is durable.

Users publishes `users.account-created.v1` through a transactional outbox trigger. Migration `003` seeds existing accounts, and the Users connector takes an initial snapshot before streaming inserts. The topic is compacted by user ID with unlimited retention and one partition. Checkout stores account facts and the next Kafka offset atomically in one non-expiring Redis hash; its direct consumer resumes from that offset and reloads it after Redis loss or rollback. Missing account facts refuse new submission with `UNAVAILABLE` until replay catches up; no Users RPC or browser-supplied timestamp is used. Operator-managed Redis uses a persistent volume, AOF, and `noeviction`.

Web forwards the internal webhook's raw bytes and `Gateway-Signature` to Checkout. Checkout verifies HMAC-SHA256 and a five-minute timestamp tolerance before accessing state, then deduplicates `gateway:<event_id>` transactionally. The payment deadline is the gateway expiry plus `PAYMENT_GATEWAY_DEADLINE_GRACE` (default two minutes). A durable status-read marker prevents repeated deadline calls; unresolved outcomes preserve stock and remain eligible for webhooks. There is no payment cancellation API.

`PAYMENT_GATEWAY_URL`, `PAYMENT_GATEWAY_API_KEY`, `PAYMENT_GATEWAY_WEBHOOK_SECRET`, `CHECKOUT_RETURN_BASE_URL` and `REDIS_ADDR` configure Checkout. An absent gateway URL refuses submission before any order or reservation is created. Gateway configuration is production-only; credentials belong only to Checkout. Deployment and Kubernetes end-to-end acceptance remain deferred.
