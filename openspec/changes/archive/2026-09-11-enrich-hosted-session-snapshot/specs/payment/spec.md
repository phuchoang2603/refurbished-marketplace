## MODIFIED Requirements

### Requirement: Payment creates hosted payment sessions by order identifier

The payment service MUST create a hosted payment session using `order_id` as the idempotency anchor and MUST return hosted-session metadata that the web edge can use to redirect the buyer.

#### Scenario: Hosted payment session is requested for a new order

- **WHEN** the web edge requests a hosted payment session for an order with buyer id, merchant id, amount, shipping address, and return context
- **THEN** the payment service SHALL persist the hosted session and payment transaction and return session metadata including `order_id`, `payment_session_id`, and the return URL

#### Scenario: Hosted payment session is requested again for the same order

- **WHEN** the web edge repeats the hosted payment session request for an order that already has a stored PENDING session
- **THEN** the payment service SHALL return stored session metadata for that `order_id` instead of creating a second independent payment identity

### Requirement: Payment snapshots commerce facts when creating a hosted session

The payment service MUST persist nested buyer and merchant snapshots (marketplace ids, optional buyer email), amount, currency, shipping address, and line items (product id, name, quantity, unit price) on hosted-session create so a later fraud gateway can score from stored commerce facts without waiting for Kafka. The payment service MUST NOT call the users service to hydrate party fields.

#### Scenario: Session create includes charge and party facts

- **WHEN** the web edge requests a hosted payment session with `order_id`, nested buyer and merchant ids, total cents, currency, shipping address, named line items, and return URL
- **THEN** the payment service SHALL store those facts with the session and SHALL create the order payment transaction in the same operation

#### Scenario: Session create omits required commerce facts

- **WHEN** the web edge requests a hosted payment session without a usable shipping address (at least line1, city, postal code, and country) or without buyer or merchant id
- **THEN** the payment service SHALL reject the request and SHALL NOT create a session
