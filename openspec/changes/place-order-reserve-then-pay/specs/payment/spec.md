## ADDED Requirements

### Requirement: Hosted payment session can be refreshed when expired and unpaid

The payment service MUST allow a repeated hosted-session request for an unpaid order whose previous session is expired or otherwise unusable to produce a new or renewed session the web edge can redirect to, without creating a second order.

#### Scenario: Session is requested again while still pending

- **WHEN** the web edge repeats the hosted payment session request for an unpaid order whose session is still pending and unexpired
- **THEN** the service SHALL return the existing session metadata instead of creating a duplicate pending session

#### Scenario: Session is requested again after expiry while unpaid

- **WHEN** the web edge requests a hosted payment session for an unpaid order whose previous session is expired
- **THEN** the service SHALL return usable hosted-session metadata for that same `order_id` so checkout can redirect again

## MODIFIED Requirements

### Requirement: Payment creates hosted payment sessions by order identifier

The payment service MUST create or reuse a hosted payment session using `order_id` as the idempotency anchor and MUST return hosted-session metadata that the web edge can use to redirect the buyer.

#### Scenario: Hosted payment session is requested for a new order

- **WHEN** the web edge requests a hosted payment session for an order with buyer, optional shipping, and return context
- **THEN** the payment service SHALL persist the hosted session state and return session metadata including `order_id`, `payment_session_id`, and the return URL

#### Scenario: Hosted payment session is requested again for the same order

- **WHEN** the web edge repeats the hosted payment session request for an order that already has a stored session
- **THEN** the payment service SHALL return stored or refreshed session metadata for that `order_id` instead of creating a second independent payment identity
