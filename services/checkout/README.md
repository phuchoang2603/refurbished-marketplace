# Checkout workflow

Checkout persists the buyer's accepted, merchant-scoped commerce snapshot and is the only coordinator for new checkouts. Orders owns the order, Inventory owns stock and reservations, and Payment owns the hosted session and payment outcome. A browser disconnect does not cancel an accepted workflow.

| Checkout state      | Allowed correlated result         | Durable command or transition                                 | Owner                |
| ------------------- | --------------------------------- | ------------------------------------------------------------- | -------------------- |
| ACCEPTED            | Accepted intent                   | Request order creation; CREATING_ORDER                        | Checkout → Orders    |
| CREATING_ORDER      | OrderCreated                      | Request reservation; RESERVING_STOCK                          | Checkout → Inventory |
| CREATING_ORDER      | OrderRejected                     | Record definitive failure; FAILED                             | Checkout             |
| CREATING_ORDER      | Deadline, result unknown          | Retry the same order operation; stay nonterminal              | Checkout/Orders      |
| RESERVING_STOCK     | StockReserved                     | Request hosted session; CREATING_SESSION                      | Checkout → Payment   |
| RESERVING_STOCK     | StockRejected or deadline         | Cancel reservation with an order-level fence; RELEASING_STOCK | Checkout → Inventory |
| CREATING_SESSION    | PaymentSessionReady               | Expose payment destination; READY_TO_PAY                      | Checkout             |
| CREATING_SESSION    | PaymentSessionRejected            | Cancel reservation; RELEASING_STOCK                           | Checkout → Inventory |
| CREATING_SESSION    | Deadline, result unknown          | Request definitive payment cancellation; RECONCILING          | Checkout → Payment   |
| READY_TO_PAY        | PaymentSucceeded                  | Commit reserved stock; COMMITTING_STOCK                       | Checkout → Inventory |
| READY_TO_PAY        | PaymentFailed or PaymentCancelled | Release stock; RELEASING_STOCK                                | Checkout → Inventory |
| READY_TO_PAY        | Deadline, outcome unknown         | Request payment verification/cancellation; RECONCILING        | Checkout → Payment   |
| RECONCILING         | Definitive no-capture result      | Release stock; RELEASING_STOCK                                | Checkout → Inventory |
| RECONCILING         | Definitive paid result            | Commit reserved stock; COMMITTING_STOCK                       | Checkout → Inventory |
| RECONCILING         | PaymentUncertain or no answer     | Preserve stock, alert; NEEDS_REVIEW                           | Checkout             |
| COMMITTING_STOCK    | StockCommitted                    | Finalize order paid; FINALIZING_PAID                          | Checkout → Orders    |
| RELEASING_STOCK     | StockReleased or absent           | Finalize order failed; FINALIZING_FAILED                      | Checkout → Orders    |
| FINALIZING_PAID     | OrderFinalized(PAID)              | COMPLETED                                                     | Checkout             |
| FINALIZING_FAILED   | OrderFinalized(FAILED)            | FAILED                                                        | Checkout             |
| COMPLETED or FAILED | Duplicate result                  | Acknowledge without replaying side effects                    | Checkout             |
| COMPLETED or FAILED | Contradictory verified payment    | Record financial exception; NEEDS_REVIEW                      | Checkout/Payment     |

Every message has a `checkout_id`, `message_id`, `operation_id`, `causation_id`, and a workflow version. Consumers must check checkout ID, order/session identity, operation, and current state; a stale result cannot roll back the workflow. A lost acknowledgement repeats the same operation, not a new order or payment attempt. Stock cancellation persists a tombstone even if it precedes reserve, so a late reserve cannot create an orphan hold. Each service persists inbox, domain changes, and outbox in one local transaction.

The order, reservation, session-creation, and payment stages have independent durable deadlines. A deadline with an unknown payment outcome must not be interpreted as failure or used to release stock until no capture is confirmed. Settlement and finalization retry until acknowledged or surfaced for investigation.
