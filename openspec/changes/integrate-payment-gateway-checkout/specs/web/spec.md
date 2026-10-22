## ADDED Requirements

### Requirement: Web forwards gateway webhooks unchanged

Web MUST expose `POST /webhooks/payment-gateway` on its in-cluster Service without browser authentication and forward the unmodified request body and `Gateway-Signature` header to Checkout. It MUST return 2xx only after Checkout accepted or acknowledged the event, 4xx when Checkout rejects the signature, and 5xx when Checkout is unavailable so the gateway retries. Web MUST NOT parse the event to change Checkout, Orders, Inventory or the cart, and MUST bound request body size. The route MUST NOT be exposed through the public Gateway.

#### Scenario: Valid webhook

- **WHEN** the gateway posts a correctly signed event
- **THEN** web forwards it to Checkout and returns 2xx after Checkout records or acknowledges it

#### Scenario: Checkout unavailable

- **WHEN** Checkout cannot be reached while a webhook is forwarded
- **THEN** web returns 5xx and does not change any cart or order state

## MODIFIED Requirements

### Requirement: Web supports authenticated seller product creation

The web service MUST expose protected browser routes that let an authenticated user create a product with explicit initial stock, a category and a condition grade through HTML form submission.

#### Scenario: Authenticated seller opens the create-product page

- **WHEN** an authenticated browser requests the seller product creation page
- **THEN** the web service SHALL render a server-side HTML page for that form

#### Scenario: Unauthenticated browser requests a seller management route

- **WHEN** an unauthenticated browser requests a protected seller product management route
- **THEN** the web service SHALL reject the request through the existing protected-route browser response instead of invoking downstream create operations

#### Scenario: Authenticated seller submits a product form

- **WHEN** an authenticated browser submits valid product details, category, condition and an initial quantity
- **THEN** the web service SHALL call products with catalog details, category, condition and explicit initial quantity, and after the listing and creation event commit SHALL report listing creation with availability processing; it SHALL NOT call EnsureStock or compensating-delete the listing

#### Scenario: Seller creation dependencies are unavailable

- **WHEN** a seller product creation request reaches the web service and products cannot durably persist the listing and creation event
- **THEN** the web service SHALL return a browser-friendly error response that matches the current popup-or-fragment mutation conventions

### Requirement: Web re-validates cart products in one batch at checkout

Web MUST re-read selected merchant-group products through a single products batch call before submitting Checkout's intent and MUST use authoritative prices, categories and conditions rather than cart snapshot values.

#### Scenario: Buyer checks out one merchant group

- **WHEN** a buyer submits checkout for a merchant group with one or more lines within the products batch size limit
- **THEN** web SHALL batch-read those products once, SHALL fail closed on missing products or wrong merchant ids, and SHALL submit an authoritative price, quantity, category and condition snapshot without removing cart lines

#### Scenario: Product missing at checkout re-validation

- **WHEN** batch lookup does not return a selected cart product
- **THEN** web SHALL reject the mutation without submitting a checkout intent

#### Scenario: Product merchant no longer matches checkout group

- **WHEN** a batch product's merchant id differs from the selected checkout group
- **THEN** web SHALL reject the mutation without submitting a checkout intent

### Requirement: Web handles hosted payment return paths safely

The web service MUST provide browser routes that let a buyer return from the gateway's hosted checkout page into a usable marketplace flow without replaying checkout or creating a new gateway payment for that order.

#### Scenario: Buyer returns after hosted payment completion

- **WHEN** the hosted checkout page redirects the browser to the return URL after a successful payment
- **THEN** the web service SHALL render the order page from Checkout's status without issuing another checkout mutation or gateway payment create

#### Scenario: Buyer returns after canceling hosted payment

- **WHEN** the buyer leaves the hosted checkout page without paying and navigates back to the marketplace
- **THEN** the web service SHALL redirect or render the buyer into a usable marketplace page for that order without issuing another checkout mutation or gateway payment create

### Requirement: Web exports traces and injects gRPC context

The web service SHALL export OpenTelemetry spans to the platform agent and inject W3C trace context on outgoing gRPC calls used for browser and gateway webhook flows so downstream services continue the same TraceId. HTTP server span names SHALL use route patterns rather than the middleware operation string or service name alone.

#### Scenario: Outgoing gRPC calls carry traceparent

- **WHEN** web invokes an internal gRPC API while handling a traced request
- **THEN** the outgoing client call includes W3C trace context derived from the active span

#### Scenario: Hosted payment callback is traced

- **WHEN** web handles `POST /webhooks/payment-gateway`
- **THEN** the request produces a server span named with method and route pattern continuing into the downstream Checkout gRPC call on the same TraceId when instrumentation is enabled

#### Scenario: Span names avoid raw path cardinality

- **WHEN** a traced request matches a parameterized chi route
- **THEN** the HTTP server span name and `http.route` use the route pattern (with placeholders) rather than the raw URL path containing concrete IDs

### Requirement: Web defers cart remove until payment succeeds

Web MUST NOT require cart removal before checkout readiness. Once Checkout confirms a paid order after stock settlement, web MUST remove that order's paid merchant-group lines when a cart identity is available.

#### Scenario: Checkout POST succeeds through place-order and session create

- **WHEN** Checkout reports a ready gateway payment for a reserved order
- **THEN** web SHALL redirect to the hosted checkout URL even if that merchant's cart lines remain

#### Scenario: Hosted payment callback reports success

- **WHEN** the buyer visits the confirmed order page and a cart identity is present
- **THEN** web SHALL remove the paid product ids with a multi-remove without affecting other merchants' cart items

### Requirement: Web removes cart lines after a terminal hosted payment

The web service MUST remove the order's product IDs from the buyer cart after the gateway payment reached a terminal status (SUCCEEDED or EXPIRED) when a cart identity is available, so an expired payment does not leave those lines for a duplicate checkout of the same reservation. A declined attempt on an open payment MUST NOT remove cart lines.

#### Scenario: Buyer returns after a failed hosted payment

- **WHEN** the buyer opens the order page after the gateway payment EXPIRED and a `cart_id` cookie is present
- **THEN** the web service SHALL multi-remove that order's product IDs from the cart

## REMOVED Requirements

### Requirement: Web accepts hosted payment outcome callbacks idempotently

**Reason**: The Payment service and its unauthenticated callback are removed.

**Migration**: Use `Web forwards gateway webhooks unchanged`; Checkout verifies and deduplicates gateway webhooks.
