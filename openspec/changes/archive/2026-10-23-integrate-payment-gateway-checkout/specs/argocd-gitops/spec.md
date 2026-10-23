## ADDED Requirements

### Requirement: Checkout payment gateway configuration per environment

The production overlay SHALL configure Checkout with the in-cluster payment gateway Edge address in the `payment-gateway` namespace, the marketplace's gateway integration ID, the gateway API key and webhook signing secret from the Checkout gateway Secret, and the buyer return URL on the `shop` host. Chart defaults and the dev overlay SHALL leave the gateway disabled, rendering no gateway address or Secret reference. The chart SHALL NOT render a Payment service or payment database.

#### Scenario: Production render

- **WHEN** the production marketplace Application renders
- **THEN** the Checkout Deployment has the gateway address and Secret reference, and no `payment` Deployment, Service or CNPG Cluster exists

#### Scenario: Dev render

- **WHEN** the dev marketplace Application renders
- **THEN** Checkout has no gateway address or Secret configuration

## REMOVED Requirements

### Requirement: Payment gateway simulator in marketplace chart

**Reason**: The simulator is replaced by the external payment gateway's hosted checkout page.

**Migration**: Remove the `payment-gateway-simulator` workload and route; use `Checkout payment gateway configuration per environment`.

### Requirement: Hosted payment URL uses edge route

**Reason**: Checkout receives the hosted checkout URL from the gateway per payment, so web no longer builds it from `HOSTED_PAYMENT_BASE_URL`.

**Migration**: Drop `HOSTED_PAYMENT_BASE_URL` and `HOSTED_PAYMENT_CALLBACK_BASE_URL` from web values.
