# Web

## Purpose

The web capability defines the public browser edge, including browser routes, auth cookie handling, and delegation to internal services.

## Requirements

### Requirement: Web owns the public browser edge

The web service MUST own the public browser surface, authorization boundary, browser auth cookies, and browser-facing UI routes, and it MUST organize those routes so public, authenticated, and non-browser concerns can apply middleware consistently while keeping unrelated browser routes available when an individual downstream domain service is unavailable.

#### Scenario: A protected route is called

- **WHEN** a client calls a protected browser endpoint
- **THEN** the web service SHALL validate the browser auth cookie and forward trusted identity to internal services

#### Scenario: A browser form is submitted

- **WHEN** a browser submits a web UI form
- **THEN** the web service SHALL process the form at the browser edge and translate successful actions into internal gRPC calls

#### Scenario: A non-browser route is called

- **WHEN** a client calls a health or simulator webhook route
- **THEN** the web service SHALL keep that route outside browser-auth middleware and preserve its documented non-browser contract

#### Scenario: A browser request enters the router

- **WHEN** a browser request enters the web router
- **THEN** the web service SHALL apply request-scoped OpenTelemetry middleware at the web edge so handlers execute with tracing context available on the request

#### Scenario: A downstream service is unavailable for one feature

- **WHEN** one downstream domain service is unavailable during a browser request
- **THEN** the web service SHALL keep unrelated browser routes and the shared shell available instead of treating the whole browser edge as unavailable

### Requirement: Web delegates auth session logic to users

The web service MUST delegate login and logout session logic to the users service while presenting browser-friendly auth UI responses and managing browser cookie persistence.

#### Scenario: Login request arrives

- **WHEN** a client calls the login endpoint
- **THEN** the web service SHALL invoke the users service for session issuance

#### Scenario: Login form is submitted from the browser

- **WHEN** a browser submits the login form
- **THEN** the web service SHALL set auth cookies and return an HTML page, fragment, or redirect that moves the user into a usable marketplace flow instead of a token-debug landing page

#### Scenario: Protected page redirects through login

- **WHEN** an unauthenticated browser is redirected away from a protected `GET` route
- **THEN** the web service SHALL preserve the intended destination and return the user to that page after successful login when it is safe to do so

#### Scenario: Protected mutation redirects through login

- **WHEN** an unauthenticated browser is redirected away from a protected `POST` route such as checkout
- **THEN** the web service SHALL return the user to a safe resume page such as `/cart` after successful login instead of replaying the mutation automatically

#### Scenario: Logout form is submitted from the browser

- **WHEN** a browser submits the logout form
- **THEN** the web service SHALL delegate token revocation to the users service, clear browser auth cookies, and return an HTML page, fragment, or redirect that leaves the browser in a usable signed-out state

### Requirement: Web supports authenticated seller product creation

The web service MUST expose protected browser routes that let an authenticated user create a product with explicit initial stock through HTML form submission.

#### Scenario: Authenticated seller opens the create-product page

- **WHEN** an authenticated browser requests the seller product creation page
- **THEN** the web service SHALL render a server-side HTML page for that form

#### Scenario: Unauthenticated browser requests a seller management route

- **WHEN** an unauthenticated browser requests a protected seller product management route
- **THEN** the web service SHALL reject the request through the existing protected-route browser response instead of invoking downstream create operations

#### Scenario: Authenticated seller submits a product form

- **WHEN** an authenticated browser submits valid product details and an initial quantity
- **THEN** the web service SHALL call the products service to create the catalog record and initial stock through the unified catalog boundary before returning a usable success response

#### Scenario: Seller creation dependencies are unavailable

- **WHEN** a seller product creation request reaches the web service and a required downstream service is unavailable
- **THEN** the web service SHALL return a browser-friendly error response that matches the current popup-or-fragment mutation conventions

### Requirement: Web keeps checkout scoped to one merchant group

The web service MUST keep checkout scoped to one merchant group per submit when a cart contains items from multiple merchants, and after creating that merchant-scoped order it MUST initiate a hosted payment session for that order and redirect the buyer to the gateway payment page.

#### Scenario: Buyer checks out one merchant group from the cart

- **WHEN** a buyer submits checkout for a selected merchant group in the cart
- **THEN** the web service SHALL create one order for only that merchant's items, leave items from other merchants in the cart, request a hosted payment session for the created order, and redirect the browser to the hosted payment URL

### Requirement: Web maps seller identity to merchant ownership in v1

The web service MUST use the authenticated user's ID as the `merchant_id` value when creating seller-owned products until a separate merchant identity model exists.

#### Scenario: Seller creates a product in v1

- **WHEN** the web service handles product creation for an authenticated user
- **THEN** it SHALL submit that user's ID as the `merchant_id` in the product creation request

#### Scenario: Seller submits product creation without explicit stock

- **WHEN** an authenticated browser submits product creation without an explicit initial stock value
- **THEN** the web service SHALL reject the request using the current browser-friendly validation or error response path instead of relying on implicit defaults

### Requirement: Web handles hosted payment return paths safely

The web service MUST provide browser routes that let a buyer return from the hosted payment gateway into a usable marketplace flow without replaying checkout.

#### Scenario: Buyer returns after hosted payment completion

- **WHEN** the hosted payment gateway redirects the browser back after a successful or failed payment attempt
- **THEN** the web service SHALL redirect or render the buyer into a usable marketplace page for that order without issuing another checkout mutation

#### Scenario: Buyer returns after canceling hosted payment

- **WHEN** the hosted payment gateway redirects the browser back after the buyer cancels payment
- **THEN** the web service SHALL return the buyer to a usable order-related page without creating a second order or payment session

### Requirement: Web accepts hosted payment outcome callbacks idempotently

The web service MUST preserve a non-browser callback path for hosted payment outcomes and MUST handle repeated callbacks safely.

#### Scenario: Gateway posts a payment outcome callback

- **WHEN** the hosted payment gateway posts an order-level payment outcome callback
- **THEN** the web service SHALL process the outcome through the payment and order path without requiring browser authentication

#### Scenario: Gateway retries a payment outcome callback

- **WHEN** the hosted payment gateway repeats the same terminal callback for an order
- **THEN** the web service SHALL handle the duplicate safely without creating a second payment result or replaying checkout behavior
