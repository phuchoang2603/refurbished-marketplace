## MODIFIED Requirements

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit, and it MUST treat checkout as a retry-safe reconciliation flow that ensures one merchant-scoped order and one hosted payment continuation for a buyer's checkout intent before redirecting the browser to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart with a new checkout intent
- **THEN** the web service SHALL create or recover one order for only that merchant's items, SHALL ensure a hosted payment session for that order, and SHALL redirect the browser to the hosted payment URL

#### Scenario: Merchant group exceeds products batch size at checkout

- **WHEN** a buyer submits checkout for a merchant group with more distinct product lines than the products batch lookup limit (100)
- **THEN** the web service SHALL reject checkout with a clear browser-facing error before calling products batch or creating an order

## ADDED Requirements

### Requirement: Web retries checkout safely

The web service MUST preserve a stable checkout-intent identifier across buyer retries for the same browser submit and MUST reconcile downstream order and payment state instead of assuming each checkout POST is a brand-new purchase attempt.

#### Scenario: Retry happens after order creation but before cart cleanup

- **WHEN** the buyer retries a checkout submit after the original request created the order but failed before cart cleanup completed
- **THEN** the web service SHALL reuse the same checkout-intent identifier, SHALL recover the existing order, and SHALL continue toward hosted payment without creating a second order

#### Scenario: Retry happens after cart cleanup but before hosted payment redirect

- **WHEN** the buyer retries checkout after the original request removed cart items or otherwise left the cart unusable for replay but did not complete hosted payment setup
- **THEN** the web service SHALL reconcile the existing order for that checkout intent, SHALL ensure hosted payment continuation for that order, and SHALL return the buyer to a usable payment path without requiring a second order

#### Scenario: Cart cleanup fails after durable checkout state exists

- **WHEN** the web service has already ensured the order and hosted payment continuation for a checkout intent but Redis cart cleanup fails
- **THEN** the web service SHALL treat cart cleanup as a non-critical side effect and SHALL NOT fail the buyer solely because the ephemeral cart mutation did not complete
