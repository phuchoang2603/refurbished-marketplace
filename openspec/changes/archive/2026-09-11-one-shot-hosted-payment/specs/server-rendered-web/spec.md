## ADDED Requirements

### Requirement: Order detail does not offer resume payment

The web service MUST NOT render a Resume payment action on the order page. Failed or expired payment MUST be shown as terminal for that order.

#### Scenario: Unpaid pending order with no successful payment

- **WHEN** the web service renders an order that is still PENDING because hosted payment has not succeeded
- **THEN** the page SHALL NOT include a control that starts hosted payment for that `order_id`

#### Scenario: Order after failed or expired hosted payment

- **WHEN** the web service renders an order whose hosted session is FAILED or EXPIRED
- **THEN** the page SHALL show that payment status and SHALL NOT offer a retry on the same order
