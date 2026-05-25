## 1. Payment Contract

- [ ] 1.1 Replace the current token-oriented payment initiation protobuf and client contract with hosted payment session creation keyed by `order_id`
- [ ] 1.2 Add the payment-domain callback contract for hosted gateway terminal outcomes and ensure the contract is idempotent by order or session identity
- [ ] 1.3 Update payment persistence models and queries for hosted session state, redirect metadata, and callback-driven terminal outcomes

## 2. Payment Service

- [ ] 2.1 Implement payment-service hosted session creation or reuse by `order_id` and return a hosted payment URL
- [ ] 2.2 Implement `services/payment` gateway callback handling that updates payment state and emits downstream order-level payment results once
- [ ] 2.3 Retire or replace the old marketplace-submitted `payment_token` initiation path so the hosted flow is the supported browser-facing contract

## 3. Dev Simulator

- [ ] 3.1 Add a dev-only hosted payment simulator under `tools/` that can render a mock payment page for a hosted session
- [ ] 3.2 Make the simulator post terminal payment callbacks to `services/payment` and then redirect the browser back to the marketplace return URL

## 4. Web Checkout And Return Flow

- [ ] 4.1 Update the merchant-scoped cart checkout handler to create the order, request a hosted payment session, and redirect the browser to the gateway URL
- [ ] 4.2 Add browser return or cancel handling that lands the buyer back on the existing `/orders/{id}` page without replaying checkout
- [ ] 4.3 Keep hosted payment outcome callbacks out of `services/web` and update web-side order rendering to reflect pending versus terminal payment state from existing domain reads

## 5. Browser Rendering

- [ ] 5.1 Update cart and order browser views to match the hosted payment flow instead of implying local payment entry on the marketplace site
- [ ] 5.2 Render a usable post-payment return state for buyers coming back from the hosted gateway

## 6. Documentation And Verification

- [ ] 6.1 Update stable architecture docs such as `docs/order-placement.md` to reflect hosted payment session creation, callback-driven outcomes, and the dev-only simulator role
- [ ] 6.2 Add or update focused tests for hosted checkout redirect behavior, return handling, simulator-driven callback flow, and duplicate callback safety in `services/web` and `services/payment`
- [ ] 6.3 Run the relevant service test suites and address any failures needed for this change
