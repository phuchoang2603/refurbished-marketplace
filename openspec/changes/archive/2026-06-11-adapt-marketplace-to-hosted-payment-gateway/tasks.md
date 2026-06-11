## 1. Payment Contract

- [x] 1.1 Replace the current token-oriented payment initiation protobuf and client contract with hosted payment session creation keyed by `order_id`
- [x] 1.2 Add the payment-domain callback contract for hosted gateway terminal outcomes and ensure the contract is idempotent by order or session identity
- [x] 1.3 Update payment persistence models and queries for hosted session state, redirect metadata, and callback-driven terminal outcomes

## 2. Payment Service

- [x] 2.1 Implement payment-service hosted session creation or reuse by `order_id` and return hosted-session metadata for the web edge to build the redirect URL
- [x] 2.2 Implement `services/payment` gateway outcome handling over gRPC that updates payment state and emits downstream order-level payment results once
- [x] 2.3 Retire or replace the old marketplace-submitted `payment_token` initiation path so the hosted flow is the supported browser-facing contract

## 3. Dev Simulator

- [x] 3.1 Add a dev-only hosted payment simulator under `tools/` that can render a mock payment page for a hosted session
- [x] 3.2 Make the simulator post terminal payment callbacks to `services/web` and then redirect the browser back to the marketplace return URL

## 4. Web Checkout And Return Flow

- [x] 4.1 Update the merchant-scoped cart checkout handler to create the order, request a hosted payment session, build the gateway URL, and redirect the browser
- [x] 4.2 Add browser return or cancel handling that lands the buyer back on the existing `/orders/{id}` page without replaying checkout
- [x] 4.3 Accept hosted payment outcome callbacks at the web edge, forward them to `services/payment` over gRPC, and render order-page payment state from existing domain reads

## 5. Browser Rendering

- [x] 5.1 Update cart and order browser views to match the hosted payment flow instead of implying local payment entry on the marketplace site
- [x] 5.2 Render a usable post-payment return state for buyers coming back from the hosted gateway

## 6. Documentation And Verification

- [x] 6.1 Update stable architecture docs such as `docs/order-placement.md` to reflect hosted payment session creation, callback-driven outcomes, and the dev-only simulator role
- [x] 6.2 Add or update focused tests for hosted checkout redirect behavior, return handling, simulator-driven callback flow, and duplicate callback safety in `services/web` and `services/payment`
- [x] 6.3 Run the relevant service test suites and address any failures needed for this change
