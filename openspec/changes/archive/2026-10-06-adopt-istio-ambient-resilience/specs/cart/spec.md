## MODIFIED Requirements

### Requirement: Cart state is ephemeral

The cart service MUST store only session/cart state and MUST NOT persist cart data in PostgreSQL. The Web edge SHALL remove purchased merchant-group lines only after the saga has finalized the order as PAID, not merely after receiving a gateway callback; repeated reads of the paid order SHALL be safe.

#### Scenario: Cart is loaded

- **WHEN** a client loads a cart
- **THEN** the service SHALL read cart state from Redis or Valkey

#### Scenario: Cart lines are removed after payment succeeds

- **WHEN** the merchant-scoped checkout completes and the order is PAID and the caller requests multi-remove of that order's product IDs
- **THEN** the service SHALL remove those product IDs from the cart document when they are still present

#### Scenario: Order is placed but unpaid

- **WHEN** an order is placed successfully and hosted payment has not completed as a paid order
- **THEN** the service SHALL NOT require those product IDs to already be absent from the cart

#### Scenario: Cart is cleared

- **WHEN** an order is finalized as PAID and the caller requests removal of its product lines
- **THEN** the service SHALL remove those paid lines while preserving other cart items

#### Scenario: Callback fails or is replayed

- **WHEN** a callback reports FAILED or EXPIRED, or a terminal callback is delivered again while settlement is pending
- **THEN** Web SHALL NOT remove cart lines before a paid order exists
