# orders Specification

## Purpose

Define merchant-scoped order persistence, buyer intent idempotency, and order lifecycle events.

## Requirements

### Requirement: Orders finalize checkout only on saga commands

Orders MUST finalize checkout orders only on correlated commands from Checkout after stock settlement. Failed finalization MUST durably fence creation even when the preparation command has not yet created an order, and SHALL acknowledge a failed order or fenced absence. Orders MUST NOT independently finalize checkout orders from inventory reservation or payment outcome events.

#### Scenario: Stock was committed after payment

- **WHEN** Checkout commands paid finalization after Inventory confirms stock commit
- **THEN** Orders SHALL update the pending order to paid and emit a correlated finalization result

#### Scenario: Stock was released after definitive failure

- **WHEN** Checkout commands failed finalization after Inventory confirms stock release or absence
- **THEN** Orders SHALL update the pending order to failed and emit a correlated finalization result

#### Scenario: Conflicting finalization command arrives

- **WHEN** Orders receives a duplicate or contradictory finalization for an already terminal order
- **THEN** Orders SHALL acknowledge exact duplicates without mutation and reject conflicting outcomes for reconciliation

#### Scenario: Failed finalization precedes preparation

- **WHEN** failed finalization arrives before the shared preparation command creates an order
- **THEN** Orders SHALL atomically persist a checkout/order-bound absence fence and a correlated finalization result, and a later preparation command SHALL NOT create a pending order

#### Scenario: Failed finalization races creation on different replicas

- **WHEN** preparation and failed finalization execute concurrently on different Orders replicas
- **THEN** Orders SHALL serialize the decisions by checkout/order identity and commit either a failed order or fenced absence without leaving a pending order

### Requirement: Orders are merchant-scoped records

The orders capability MUST persist each created order with exactly one caller-supplied `merchant_id` and MUST accept only merchant-scoped order creation requests.

#### Scenario: Single-merchant order is created

- **WHEN** a caller creates an order with a merchant identifier and items that all belong to that merchant grouping
- **THEN** the orders service SHALL persist one order record for that merchant and the corresponding order items

#### Scenario: Merchant identifier is omitted from order creation

- **WHEN** a caller submits an order creation request without a merchant identifier
- **THEN** the orders service SHALL reject the request as invalid

#### Scenario: Mixed-merchant checkout is handled upstream

- **WHEN** a checkout flow contains items from more than one merchant
- **THEN** the upstream caller SHALL split that work into separate merchant-scoped order creation requests before calling the orders service

### Requirement: Orders write order-level outbox events

The orders capability MUST write one outbox row per created order, not one outbox row per order item, and the correlated result MUST identify the checkout, order, and preparation attempt. Checkout order creation MUST be driven by Checkout's shared preparation command in an Orders consumer group distinct from Inventory's group. Inventory preparation SHALL NOT depend on publication or observation of the order-created result.

#### Scenario: Merchant order is created

- **WHEN** an order is persisted successfully in response to Checkout's shared preparation command
- **THEN** the service SHALL store the order, its items, and one correlated order-created outbox row in the same transaction

#### Scenario: Order-created payload is published

- **WHEN** the service writes the order-created outbox row
- **THEN** the result SHALL carry checkout and order identifiers and the correlated preparation operation and version; consumers SHALL obtain preparation item facts from the shared command rather than infer them from the order-created result

#### Scenario: Place-order returns before Kafka consume

- **WHEN** Orders has persisted a new checkout order but its order-created event is delayed
- **THEN** Inventory SHALL be able to reserve independently, while Checkout SHALL withhold gateway payment until both successful preparation results are durably observed

### Requirement: Place order is idempotent per buyer intent

The orders capability MUST accept a caller-supplied `idempotency_key` on place-order and MUST persist it with a uniqueness constraint per buyer so retries return the original order instead of inserting a second row.

#### Scenario: First place-order for an intent

- **WHEN** a caller places an order with a new `idempotency_key` for that buyer
- **THEN** the service SHALL persist one order and associated items for that key

#### Scenario: Retry with the same buyer and key

- **WHEN** the same buyer places an order again with the same `idempotency_key`
- **THEN** the service SHALL return the existing order and SHALL NOT insert a second order row

#### Scenario: Same key with a conflicting body

- **WHEN** the same buyer reuses an `idempotency_key` with a different merchant or different item lines
- **THEN** the service SHALL reject the request as a conflict without creating another order

#### Scenario: Idempotency key is omitted

- **WHEN** a caller places an order without an `idempotency_key`
- **THEN** the service SHALL reject the request as invalid

### Requirement: Place order persists without calling products

Orders MUST persist merchant-scoped orders and emit a correlated order-created result without calling Products or Inventory. Orders MUST consume the shared preparation command without waiting for Inventory's outcome. Deterministic business rejection MUST produce a durable correlated order-rejected result; transient infrastructure failures MUST remain retryable.

#### Scenario: Stock is available

- **WHEN** Orders processes a checkout order command
- **THEN** Orders SHALL persist the pending order and report its creation without synchronously reserving stock or creating payment

#### Scenario: Retry after successful persist

- **WHEN** the same checkout order command is retried after the order was created
- **THEN** Orders SHALL return or re-emit the original correlated outcome without inserting another order

#### Scenario: Preparation is rejected

- **WHEN** a well-formed correlated preparation command fails deterministic order business validation
- **THEN** Orders SHALL persist its definitive rejected intent and correlated rejection together without creating a pending order, and redelivery SHALL preserve that outcome
