## ADDED Requirements

### Requirement: Web stamps cart line product snapshots

The web service MUST obtain product name and unit price from the products service when adding or updating a cart line and MUST write those values onto the cart item so subsequent cart reads do not require per-line product hydration.

#### Scenario: Buyer adds a product to the cart

- **WHEN** a browser adds a cart item with product_id, merchant_id, and quantity
- **THEN** the web service SHALL load that product once from products, SHALL reject merchant mismatch against the product record, and SHALL call cart add with product name and unit price snapshot fields

#### Scenario: Buyer changes cart line quantity

- **WHEN** a browser sets quantity for an existing cart line (quantity greater than zero)
- **THEN** the web service SHALL refresh the product snapshot from products and pass it into cart set-quantity

### Requirement: Web renders cart from stored snapshots

The web service MUST build cart HTML from cart item snapshot fields without issuing per-line product gets for cart page and cart fragment re-renders.

#### Scenario: Cart page is loaded

- **WHEN** a browser opens the cart page or receives a cart fragment after a cart mutation
- **THEN** the web service SHALL render names, unit prices, and line totals from the cart payload and SHALL NOT call products once per cart line

### Requirement: Web re-validates cart products in one batch at checkout

The web service MUST re-read selected merchant group products through a single batch products API before creating an order and MUST use those authoritative prices for order lines rather than cart snapshot prices.

#### Scenario: Buyer checks out one merchant group

- **WHEN** a buyer submits checkout for a merchant group containing one or more cart lines
- **THEN** the web service SHALL call products once with all selected product IDs for that group, SHALL fail closed if any selected product is missing or has a different merchant_id, SHALL create the order using batch-returned unit prices, and SHALL remove those product IDs from the cart with a single multi-remove call rather than N single removes

#### Scenario: Product missing at checkout re-validation

- **WHEN** the batch product lookup does not return a product referenced by a selected cart line
- **THEN** the web service SHALL fail the checkout mutation without creating an order for that submit

#### Scenario: Product merchant no longer matches checkout group

- **WHEN** the batch product lookup returns a product whose merchant_id differs from the selected checkout merchant group
- **THEN** the web service SHALL fail the checkout mutation without creating an order for that submit
