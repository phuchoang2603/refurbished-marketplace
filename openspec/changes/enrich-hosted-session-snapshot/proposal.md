## Why

Hosted-session create already snapshots merchant, amount, and line ids so a later fraud gateway can score without waiting on Kafka, but shipping is empty and lines have no names. The simulator stays a dummy; checkout should send the same kind of commerce facts a PSP expects on PaymentIntent create (customer id, connected-account id, ship-to, line items).

## What Changes

- **BREAKING (payment gRPC):** `CreateHostedPaymentSession` uses nested buyer and merchant snapshots (marketplace ids), requires a shipping address, and stores product names on line items.
- Checkout collects a postal shipping address and sends party ids, optional buyer email from the access-token `eml` claim, shipping, and named lines. No users gRPC on checkout or payment create.
- Payment persists that nested snapshot on the intent for a future gateway POST. Do not put PII on the hosted-pay query string.
- Simulator stays three-button.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `payment`: Session create persists nested party ids, shipping, and named line items (no users hydration).
- `web`: Checkout collects shipping and passes party ids, optional JWT email, shipping, and line names.
- `server-rendered-web`: Cart/checkout UI includes shipping fields for the selected merchant group.

## Impact

- `shared/proto/payment/v1` (nested party snapshots, line `name`).
- `services/payment` persist path (buyer/merchant JSON + line names + required shipping).
- `services/web` checkout shipping form and session mapping (JWT `eml` if present).

## Non-goals

- Account `created_at`, users hops, payment→users mesh.
- Fraud scoring or replacing the simulator with a real PSP.
- Device/IP on marketplace pages; seller KYC/Connect onboarding (that lives on the PSP later).
- Billing address UI, address-book persistence, category/condition catalog fields.
- HMAC-signed webhooks or changing hosted URL query-string shape.
