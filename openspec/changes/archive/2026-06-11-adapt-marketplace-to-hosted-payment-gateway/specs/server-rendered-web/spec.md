## MODIFIED Requirements

### Requirement: Web renders merchant-grouped cart checkout

The web service MUST render cart items grouped by `merchant_id` and expose a separate checkout action per merchant group so checkout remains single-order per submit, and a successful checkout submit MUST continue the browser flow into a hosted payment redirect instead of stopping at a local payment-entry page.

#### Scenario: Cart contains items from multiple merchants

- **WHEN** the web service renders the cart page for a cart containing items from different merchants
- **THEN** the page SHALL group those items by merchant and render a merchant-scoped subtotal and checkout action for each group

#### Scenario: Merchant checkout action succeeds

- **WHEN** a buyer submits a merchant-scoped checkout action successfully
- **THEN** the browser flow SHALL continue into the hosted payment redirect for that created order instead of rendering a marketplace-hosted payment form

## ADDED Requirements

### Requirement: Web renders usable hosted payment return pages

The web service MUST render or redirect to a usable marketplace page when the buyer returns from the hosted payment gateway.

#### Scenario: Buyer returns after payment attempt

- **WHEN** the browser returns from the hosted payment gateway to the marketplace
- **THEN** the web service SHALL render or redirect to a complete HTML page that shows the buyer a usable post-payment state for the order
