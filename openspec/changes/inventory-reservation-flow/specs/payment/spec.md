## MODIFIED Requirements

### Requirement: Payment consumes order-created events

The payment service MUST consume successful inventory reservation events and create one payment transaction per order.

#### Scenario: Inventory reservation is received

- **WHEN** the service receives an inventory-reserved event for an order
- **THEN** it SHALL deduplicate the message and create or update the order payment transaction

#### Scenario: Inventory reservation fails upstream

- **WHEN** inventory emits a reservation-failed event for an order
- **THEN** the payment service SHALL NOT create a payment transaction for that order from the failed reservation path

## ADDED Requirements

### Requirement: Payment emits order-level outcome events after reservation

The payment service MUST emit order-level payment success and failure events only for orders whose inventory reservation has succeeded.

#### Scenario: Reserved order payment completes

- **WHEN** an order payment transaction for a reserved order succeeds or fails
- **THEN** the service SHALL write the corresponding order-level payment outbox event for downstream consumers
