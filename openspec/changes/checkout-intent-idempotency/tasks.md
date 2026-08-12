## 1. Orders idempotency contract

- [x] 1.1 Extend the orders protobuf and service contract to accept a buyer-scoped checkout-intent idempotency key on order creation.
- [x] 1.2 Add orders persistence support for buyer-scoped idempotency lookup and uniqueness enforcement, including conflicting payload detection for reused keys.
- [x] 1.3 Update orders service validation and response behavior so repeated creates for the same buyer intent return the existing order instead of inserting a duplicate.

## 2. Payment session reconciliation

- [x] 2.1 Update payment hosted-session behavior so repeated requests for the same unpaid order return the reusable stored session metadata.
- [x] 2.2 Add payment-session refresh behavior for expired or otherwise unusable unpaid sessions without requiring a new order.
- [x] 2.3 Keep terminal payment outcomes from reopening paid or otherwise terminal order checkout paths.

## 3. Web checkout retry safety

- [x] 3.1 Update the checkout form and browser-edge flow to generate and preserve one checkout-intent identifier per buyer submit intent.
- [x] 3.2 Change the cart checkout handler to reconcile order creation and hosted payment session setup for retries instead of assuming every POST is a new checkout.
- [x] 3.3 Move Redis cart cleanup out of the critical checkout success path so durable order and payment state can continue even when cart mutation fails.
- [x] 3.4 Ensure browser return or retry flows can send the buyer back into a usable payment path for an existing pending order without creating a second order.

## 4. Docs and verification

- [x] 4.1 Update stable architecture docs such as `docs/order-placement.md` to describe edge reconciliation versus outbox or inbox-driven lifecycle choreography.
- [x] 4.2 Verify the implementation against the planned retry paths: repeated submit after order create, retry after cart cleanup, and retry after expired hosted session.
