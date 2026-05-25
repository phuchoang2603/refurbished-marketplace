## Why

The repository currently treats checkout as an internal order-creation flow followed by a redirect to an order detail page, while payment entry is modeled as an internal `payment_token` submission path that is not actually wired into the browser experience. To support a separate ecommerce fraud gateway with hosted payment entry, the marketplace needs a cleaner boundary where it sends commerce facts to an external payment service and redirects the buyer there instead of collecting payment details itself.

## What Changes

- Replace the browser checkout follow-up flow so the marketplace creates an order, requests a hosted payment session from the payment domain, and redirects the buyer to the gateway payment page.
- Redefine the payment-domain contract around hosted payment session creation keyed by `order_id` instead of marketplace-submitted `payment_token` input.
- Add a browser return and callback outcome flow so gateway payment results can move orders into usable post-payment states without replaying checkout.
- Limit the marketplace-to-gateway request payload to commerce facts needed for payment and fraud context, including order identifiers, buyer and merchant identifiers, line-item summaries, shipping address, and return URLs.
- Keep derived behavioral history and fraud features out of the marketplace runtime contract; those remain future gateway or analytics concerns.
- Do not introduce customer or merchant profile-sync event streams in this change.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `web`: checkout and payment browser behavior changes from local payment initiation assumptions to hosted gateway session creation, redirect, and callback handling.
- `server-rendered-web`: rendered checkout and order follow-up pages change so the browser is redirected into a hosted payment flow and returned to a usable marketplace page afterward.
- `payment`: payment requirements change from marketplace-supplied payment-token initiation toward order-id-keyed hosted session orchestration and gateway outcome handling.

## Impact

- Affected code will include `services/web` checkout handlers, protected routes, order detail or return routes, and shared browser response helpers.
- The payment service gRPC and storage contract will change to support hosted payment sessions and gateway callbacks instead of the current internal token-oriented initiation shape.
- A small dev-only hosted payment simulator under `tools/` may be added so the marketplace can exercise redirect, callback, and return flows before a real external gateway exists.
- Existing docs such as `docs/order-placement.md` and `docs/ecommerce-fraud-gateway.md` will become the primary architecture context for the implementation.
- Non-goals for this change include stored payment methods, customer or merchant profile synchronization over Kafka, and full fraud-feature pipeline implementation.
