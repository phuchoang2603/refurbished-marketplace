## Context

See proposal.md for motivation. Session create already writes `payment_intents` plus a payment transaction. Checkout has buyer id from the access token (`sub`) and merchant id from the cart group. JWTs also carry `eml`. Account age is out of scope, so neither web nor payment calls users at session create.

## Goals / Non-Goals

**Goals:**

- Nested party messages (ids; optional buyer email from JWT).
- Shipping from the checkout form; named lines from the products batch already used at checkout.
- Keep the hosted redirect URL free of PII.

**Non-Goals:**

- Users hops, `created_at`, payment→users mesh.
- Seller KYC / Connect onboarding in this change (PSP stores that on the connected account).
- Scoring, simulator UI, billing form, address book.

## Decisions

### 1. Nested snapshots without users hydration

```text
CreateHostedPaymentSessionRequest
  order_id, currency, total_cents, return_url
  buyer:    { id, email? }   // email from JWT eml only
  merchant: { id }           // marketplace merchant_id; PSP Connect id later
  shipping_address: Address
  items: { product_id, name, quantity, unit_price_cents }
```

Persist buyer/merchant JSON as sent. Do not fetch users. Do not copy shipping name into email.

This matches Stripe PaymentIntent create: `customer` (or customer id), `transfer_data.destination` / connected account id, `shipping`, line items. Stripe already holds seller legal entity, bank, and account age on the **Connect account** created at seller onboarding — not re-collected on each buyer checkout.

### 2. No users hops if age is not required

Buyer id is `sub`. Buyer email is optional `eml` (not a users RPC). Merchant is the cart `merchant_id` already on the order. Shipping is the form.

**Alternative:** Payment `GetUserByID` for email/age. Rejected for this change: age is not needed; Connect-style seller profile is not a checkout field.

### 3. Simulator unchanged

Do not add snapshot fields to `/pay` query params.

## Risks / Trade-offs

- [JWT without `eml` on old tokens] → Persist buyer id only; email optional.
- [Deploy skew] → Deploy payment (required shipping) then web.
- [PII] → Do not log email/address in full; do not put them on URLs.

## Migration Plan

1. Proto + intent JSON + required shipping validation.
2. Web shipping form + nested ids + optional JWT email + named lines.
3. Rollback: revert web first.

## Open Questions

None that block implementation.
