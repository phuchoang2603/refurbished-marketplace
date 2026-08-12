## MODIFIED Requirements

### Requirement: Payment creates hosted payment sessions by order identifier

The payment service MUST create, reuse, or refresh a hosted payment session using `order_id` as the durable payment-continuation anchor and MUST return hosted-session metadata that the web edge can use to redirect or reattach the buyer safely.

#### Scenario: Hosted payment session is requested for a new order

- **WHEN** the web edge requests a hosted payment session for an order with buyer, optional shipping, and return context
- **THEN** the payment service SHALL persist hosted session state and return session metadata including `order_id`, `payment_session_id`, and return or cancel URLs

#### Scenario: Hosted payment session is requested again for the same unpaid order

- **WHEN** the web edge repeats the hosted payment session request for an order that already has a reusable unpaid stored session
- **THEN** the payment service SHALL return the same stored session metadata instead of creating a duplicate session

#### Scenario: Hosted payment session is retried after the stored session expired

- **WHEN** the web edge requests a hosted payment session for an unpaid order whose stored session is expired or otherwise no longer usable for redirect
- **THEN** the payment service SHALL refresh the order's hosted payment continuation and return usable current session metadata without requiring a new order
