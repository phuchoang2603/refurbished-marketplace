## Context

The current marketplace browser flow creates an order from the cart and redirects directly to `/orders/{id}`, but the order page does not render a payment-entry experience. At the same time, the internal payment contract is still built around `InitiatePayment(order_id, buyer_user_id, payment_token, ...)`, which assumes payment details are collected by the marketplace and converted into a token before payment processing begins.

That shape does not fit the intended separation where a dedicated payment gateway owns hosted payment entry, runtime fraud signals, and eventual payment outcome callbacks. This change crosses the web and payment boundaries, alters the buyer checkout flow, and changes the payment-domain contract, so it benefits from an explicit design before implementation.

## Goals / Non-Goals

**Goals:**

- Move the buyer payment-entry experience out of the marketplace and into a hosted payment page.
- Keep the marketplace-to-gateway request limited to commerce facts that the marketplace already owns.
- Use `order_id` as the payment-session idempotency anchor.
- Add a callback or return flow so the buyer lands back in a usable marketplace page after payment.
- Leave room for a future fraud-feature pipeline without making the marketplace responsible for behavioral aggregates.

**Non-Goals:**

- Building a real external gateway service in this change.
- Adding stored payment methods, billing-address collection, or guest checkout.
- Introducing customer or merchant synchronization events just to seed gateway-local profiles.
- Implementing Spark, Flink, or any online feature store.
- Redesigning merchant-scoped order creation or mixed-merchant cart grouping.
- Putting gateway truth or long-lived payment-domain logic into a dev simulator.

## Decisions

### Decision: Use hosted payment session creation instead of marketplace-submitted payment tokens

The payment boundary will shift from marketplace-submitted `payment_token` input to hosted payment session creation keyed by `order_id`. The web edge will create an order, call the payment domain with order, buyer, and redirect context, receive hosted-session metadata, build the buyer-facing hosted payment URL locally, and redirect the browser to the configured gateway base URL.

Rationale:

- It matches the intended service split where payment entry belongs to the gateway, not the marketplace.
- It removes raw payment-entry concerns from the marketplace browser surface.
- It creates a cleaner place for device, network, and fraud-signal capture.

Alternatives considered:

- Keep marketplace-hosted payment entry and only swap the downstream processor. Rejected because it preserves the wrong browser and ownership boundary.
- Collect payment details on the order page and post a token internally. Rejected because it still makes the marketplace the payment-entry surface.

### Decision: Treat `order_id` as the session-creation idempotency key

Repeated session-creation requests for the same `order_id` should return the same active payment session instead of creating duplicates.

Rationale:

- The order already anchors buyer intent in the existing marketplace model.
- It avoids a redundant parallel `idempotency_key` field that would just restate the same business identity.
- It keeps retry behavior understandable at both the web and payment layers.

Alternatives considered:

- Add a separate caller-supplied `idempotency_key`. Rejected because the initial design would just duplicate `order_id`.

### Decision: Keep the marketplace payload narrow and commerce-owned

The marketplace request to the payment domain will include only data it already owns in v1: order ID, buyer ID, currency, optional shipping address, and return or cancel URLs. Merchant metadata and line-item summaries remain order-domain facts and are not required on the payment-session request in the first version.

Rationale:

- It keeps responsibilities clear.
- It avoids coupling the marketplace to gateway-side fraud or payment-method internals.
- It aligns with the future fraud plan where runtime signals are captured at the hosted page and behavioral aggregates are computed elsewhere.

Alternatives considered:

- Add billing address or marketplace-derived behavioral features now. Rejected because they are unnecessary for the first hosted-flow boundary and expand scope without helping the core flow.
- Add a marketplace checkout shipping-address form now. Rejected for v1 because the hosted flow does not require it; payment accepts an empty shipping address until the marketplace owns address collection.

### Decision: Create gateway-local customer and merchant records lazily, if needed

This change will not require marketplace-emitted `customer.created` or `merchant.created` synchronization events. If the payment domain needs local records, it can create or update them lazily when a payment session is created.

Rationale:

- It keeps the change focused on checkout and payment flow.
- It avoids introducing Kafka profile-sync contracts that the user explicitly does not want.
- It remains compatible with simulation work where customer and merchant profiles may be generated entirely inside the gateway project.

Alternatives considered:

- Add profile-sync events now. Rejected because they are not required for the hosted-payment flow itself.

### Decision: Preserve merchant-scoped order creation and order-first checkout

The marketplace will continue to create one merchant-scoped order per checkout submit before requesting a hosted payment session.

Rationale:

- It preserves the current cart-grouping and order-creation rules.
- It gives the hosted payment flow a stable business identifier immediately.
- It minimizes changes outside web and payment.

Alternatives considered:

- Delay order creation until after payment capture. Rejected because it conflicts with the current merchant-scoped order model and would require larger downstream contract changes.

### Decision: Terminate server-to-server payment callbacks at the web edge

Hosted gateway callbacks will terminate in `services/web` over HTTP, and the web edge will forward terminal outcomes to `services/payment` over gRPC. Browser return paths also terminate in `services/web`.

Rationale:

- It keeps `services/payment` internal and gRPC-only.
- It gives the marketplace a single public HTTP edge for both browser navigation and gateway callbacks.
- Payment truth and idempotent state transitions still live in `services/payment`; web only adapts the HTTP callback contract.

Alternatives considered:

- Terminate callbacks directly in `services/payment` over HTTP. Rejected because it forces the payment service to expose a public HTTP surface.
- Accept callbacks in both services. Rejected because it increases coordination complexity without helping the simple hosted-flow design.

### Decision: Build hosted payment redirect URLs in `services/web`

The web edge will build the buyer-facing hosted payment URL from payment-session metadata plus `HOSTED_PAYMENT_BASE_URL`. Return, cancel, and callback URLs are derived from the checkout request host.

Rationale:

- It keeps gateway and browser URL configuration at the public edge.
- It lets payment stay focused on session persistence and outcome handling.
- It avoids payment-service knowledge of simulators, port-forwards, or external gateway hosts.

Alternatives considered:

- Have `services/payment` return a fully formed hosted payment URL. Rejected because payment should not own browser-edge URL construction.

### Decision: Return buyers to the existing order page

The buyer return URL from the hosted payment page will land on the existing order detail route for the created order.

Rationale:

- It reuses an existing page and avoids introducing a new return-page surface.
- It gives the buyer a natural place to see pending, paid, or failed order state.
- It minimizes routing, template, and navigation changes in `services/web`.

Alternatives considered:

- Add a dedicated payment return page. Rejected for the first version because it adds another route and template without being required for a usable buyer flow.

### Decision: Keep inventory reservation timing unchanged in v1

The hosted payment session boundary will not change the current order-first reservation timing in the first version of this change.

Rationale:

- It preserves the existing order, inventory, and payment sequencing model.
- It keeps this change focused on hosted payment redirection and callback handling.
- It avoids a larger redesign around delayed reservation or payment-first flows.

Alternatives considered:

- Introduce a new pre-reservation pending-session state that changes when reservation happens. Rejected for v1 because it expands scope beyond the hosted payment boundary itself.

### Decision: Store only minimal hosted-session state in `services/payment`

`services/payment` will store only the smallest hosted-session record needed to support session reuse, callback correlation, and clear order-page payment status.

Minimum state:

- `order_id`
- `payment_session_id`
- `status`
- `expires_at` when the hosted session can time out
- `failure_reason` when a terminal failure exists
- `created_at` and `updated_at`

The `status` set should stay minimal in v1:

- `pending` for created or active hosted sessions awaiting buyer action
- `succeeded`
- `failed`
- `cancelled`
- `expired`

Rationale:

- It is enough for `order_id`-based idempotent session reuse.
- It is enough to correlate hosted gateway callbacks back to marketplace orders.
- It is enough for the existing `/orders/{id}` page to show pending versus terminal payment state cleanly.
- It avoids introducing unnecessary payment-method, device, or fraud-feature persistence into this change.

Alternatives considered:

- Persist a richer hosted-session model with payment-method, device, and gateway-debug metadata. Rejected for v1 because it is not required for the buyer flow or basic callback handling.
- Persist no hosted-session record and infer everything from order state alone. Rejected because the order page would not be able to distinguish pending hosted payment from terminal payment outcomes cleanly.

### Decision: Render payment state separately on the existing order page

The existing `/orders/{id}` page will remain the buyer landing page after hosted payment, and it will present a separate payment-status section driven by the hosted-session state instead of expanding the order-status model to represent every payment outcome.

Rationale:

- It keeps order lifecycle and payment lifecycle distinct.
- It avoids turning the order-status enum into a mixed order-plus-payment state machine.
- It reuses the existing order page while still letting buyers see meaningful hosted payment outcomes such as pending, failed, cancelled, and expired.

Expected v1 page behavior:

- `Order status` continues to use the current order lifecycle states.
- `Payment status` is shown as a separate section using the hosted-session status.
- `PENDING` orders can still show `pending`, `failed`, `cancelled`, or `expired` payment state without requiring a new dedicated payment-management page.
- `PAID` orders show successful payment state.

Alternatives considered:

- Expand order states to represent every payment outcome. Rejected because it mixes separate business concerns and adds avoidable state complexity.
- Add a dedicated payment return or management page. Rejected for v1 because the existing order page can show the necessary payment state with less routing and UI work.

### Decision: Use a dev-only hosted payment simulator under `tools/`

Until a real external gateway exists, the repository may include a tiny hosted payment simulator under `tools/` that renders a mock payment page, posts a payment callback to `services/payment`, and redirects the browser back to the marketplace.

Rationale:

- It exercises the real redirect, callback, and return flow without requiring a full gateway implementation.
- It keeps payment truth in `services/payment` instead of in the simulator.
- It avoids mis-framing the simulator as a reusable client library or a production service.

Alternatives considered:

- Put the simulator under `clients/`. Rejected because the simulator is not a client library and that location blurs responsibilities.
- Put full payment logic into the simulator. Rejected because it would duplicate payment-domain behavior that should stay in `services/payment`.

## Risks / Trade-offs

- [Hosted payment redirect introduces another user-visible round trip] -> Mitigate by building a direct hosted URL at the web edge and using standard browser redirects from the checkout handler.
- [Order creation before payment can leave more pending orders after buyer abandonment] -> Mitigate by keeping order status pending until callback completion and documenting follow-up cleanup or expiration work as future scope.
- [Changing the payment API from token-based initiation can break existing internal assumptions] -> Mitigate by updating the payment spec, protobuf contract, and web integration together in one scoped change.
- [Gateway callbacks can arrive more than once or after browser return paths] -> Mitigate by making payment outcome handling idempotent and order-state updates tolerant of repeated terminal events.

## Migration Plan

1. Update the payment and web contracts in OpenSpec and protobuf definitions for hosted session creation and outcome handling.
2. Implement payment-domain support for creating or reusing a hosted session by `order_id` and returning hosted-session metadata over gRPC.
3. Add a dev-only hosted payment simulator under `tools/` that can render a mock hosted page, submit a terminal callback to `services/web`, and redirect the browser back to the marketplace.
4. Update the web checkout flow to create the order, request the hosted payment session, build the gateway URL, and redirect the browser.
5. Add browser return handling and web-edge callback forwarding so order pages and order state remain usable after payment completion or cancellation.
6. Remove or retire the old marketplace-submitted token initiation path once the hosted flow is wired end to end.

Rollback strategy:

- Revert the web checkout redirect to the existing order-detail redirect path.
- Restore the prior payment initiation contract if the hosted-session path is not yet safe to serve.

## Open Questions
