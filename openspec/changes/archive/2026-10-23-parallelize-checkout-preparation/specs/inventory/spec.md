## MODIFIED Requirements

### Requirement: Inventory processes Checkout reservation commands

Inventory MUST consume Checkout's shared preparation command in an Inventory group distinct from Orders, plus correlated commit and release commands idempotently, and report each result from its own committed transaction. An `orders.created` event MUST NOT independently reserve stock for a saga checkout.

#### Scenario: Order-created event arrives without Checkout command

- **WHEN** Inventory observes an `orders.created` event for a saga checkout before a shared preparation command
- **THEN** it SHALL NOT reserve stock or emit `inventory.reserved` from that event alone

#### Scenario: Reservation command is redelivered

- **WHEN** Checkout's shared preparation command is redelivered for the same order
- **THEN** Inventory SHALL NOT double-hold stock and SHALL report the existing definitive reservation result

#### Scenario: Reservation cannot be completed

- **WHEN** one or more lines cannot be fully reserved
- **THEN** Inventory SHALL leave no partial active hold and SHALL report a correlated reservation failure

#### Scenario: Web attempts a direct checkout reserve

- **WHEN** web attempts to reserve stock for checkout using the removed synchronous path
- **THEN** the service or mesh SHALL reject the mutation and SHALL NOT create a hold

#### Scenario: Order result is delayed

- **WHEN** Inventory receives a shared preparation command before Orders creates the order or publishes its result
- **THEN** Inventory SHALL reserve from that command's complete merchant-scoped item snapshot without calling Orders or waiting for an order-created event

#### Scenario: Multiple checkouts compete for limited stock

- **WHEN** Inventory replicas process distinct checkouts for the same products concurrently across Kafka partitions
- **THEN** stock mutation SHALL remain atomic for all lines and total successful reservations SHALL NOT exceed available quantities

### Requirement: Inventory seeds stock from ProductCreated

Inventory SHALL consume ProductCreated on `products.created` in a consumer group independent of the search projector and of inventory's checkout command consumer. It SHALL validate an explicit non-negative initial quantity and atomically persist the event id in its inbox with a stock row whose reserved quantity is zero. EnsureStock SHALL remain an internal operation and SHALL NOT be exposed over gRPC.

#### Scenario: First seed succeeds

- **WHEN** inventory consumes a valid ProductCreated for an unseeded product
- **THEN** it SHALL commit the inbox record and initial stock together before acknowledging processing

#### Scenario: Duplicate seed for the same product

- **WHEN** a creation event is redelivered or replayed after the product has been seeded, including after reservations have changed its quantities
- **THEN** inventory SHALL NOT add, reset, or overwrite stock and SHALL treat a matching seed intent as successfully processed

#### Scenario: Conflicting seed intent

- **WHEN** a creation event attempts to seed an existing product with conflicting creation intent
- **THEN** inventory SHALL preserve the existing ledger and surface a processing error rather than reset quantity

#### Scenario: Missing or invalid initial quantity

- **WHEN** ProductCreated omits initial quantity or carries a negative quantity
- **THEN** inventory SHALL reject processing without seeding stock or recording successful consumption

#### Scenario: Consumer fails before commit

- **WHEN** creation-event processing fails before its database transaction commits
- **THEN** neither successful inbox consumption nor initial stock SHALL be committed, and transient failures SHALL remain retryable without deleting the catalog listing

#### Scenario: Consumer retries after commit

- **WHEN** the stock transaction committed but delivery is retried before acknowledgement completed
- **THEN** inventory SHALL recognize the committed event and SHALL NOT seed stock again

#### Scenario: Creation consumption is isolated from reservation topics

- **WHEN** ProductCreated processing returns a retryable or validation error
- **THEN** inventory's checkout preparation and stock-settlement consumer SHALL continue independently and SHALL NOT stall reservation settlement on that error
