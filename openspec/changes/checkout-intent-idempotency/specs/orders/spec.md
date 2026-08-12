## ADDED Requirements

### Requirement: Orders accept buyer-scoped checkout intent idempotency

The orders service MUST accept a caller-supplied checkout-intent idempotency key for order creation, scoped to the buyer, so repeated creates for the same buyer intent return one merchant-scoped order instead of creating duplicates.

#### Scenario: New buyer checkout intent creates an order

- **WHEN** a caller creates an order with a buyer identifier, merchant identifier, items, total amount, and a checkout-intent idempotency key that has not been used for that buyer
- **THEN** the orders service SHALL persist one new order for that buyer intent and return it

#### Scenario: Repeated create uses the same buyer checkout intent

- **WHEN** a caller repeats order creation for the same buyer and checkout-intent idempotency key with the same merchant-scoped order payload
- **THEN** the orders service SHALL return the existing order for that buyer intent and SHALL NOT create a second order row

#### Scenario: Repeated create reuses the same key with a different payload

- **WHEN** a caller repeats order creation for the same buyer and checkout-intent idempotency key but with a different merchant identifier, order items, or total amount
- **THEN** the orders service SHALL reject the request as a conflicting reuse of an existing buyer checkout intent
