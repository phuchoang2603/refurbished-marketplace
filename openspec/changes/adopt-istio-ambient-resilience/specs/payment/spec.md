## REMOVED Requirements

### Requirement: Payment expires abandoned hosted sessions

**Reason**: The only remaining hosted sessions belong to Checkout, which reconciles uncertain outcomes rather than sweeping them after a local deadline.
**Migration**: None; fresh-cluster checkout sessions use the replacement requirement below.

#### Scenario: No independent expiry

- **WHEN** a Checkout-owned hosted session passes its local deadline
- **THEN** Payment does not expire it through the retired pre-Checkout sweep

## ADDED Requirements

### Requirement: Payment reconciles Checkout-owned session expiry

Payment MUST distinguish EXPIRED from FAILED. For Checkout-owned sessions, Payment MUST record EXPIRED and report a correlated definitive failure to Checkout only when the gateway explicitly confirms expiry without capture. A local deadline alone MUST NOT mark a session EXPIRED or FAILED: Checkout requests cancellation or verification and keeps uncertain outcomes unresolved. Payment SHALL NOT run a non-Checkout automatic expiry sweep or create an unreserved hosted session.

#### Scenario: Pending session passes its deadline

- **WHEN** a PENDING Checkout session passes its expiry deadline and the gateway explicitly confirms no capture
- **THEN** Payment SHALL mark the session EXPIRED and publish a correlated definitive result with an expiry reason to Checkout; without confirmation, Payment SHALL leave the outcome uncertain for Checkout to reconcile

#### Scenario: Gateway reports unresolved outcome

- **WHEN** the gateway cannot prove that capture did not occur after a local deadline
- **THEN** Payment SHALL leave the Checkout session pending verification, report an unresolved status on cancellation, and SHALL NOT assert that no payment occurred

#### Scenario: No independent session expiry

- **WHEN** a Checkout-owned session is merely past its local deadline
- **THEN** Payment SHALL NOT expire it through an independent automatic session sweep
