## REMOVED Requirements

### Requirement: Payment emits order-level outcome events after reservation

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.

### Requirement: Payment persists inbox and outbox state

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.

### Requirement: Payment creates hosted payment sessions by order identifier

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.

### Requirement: Payment accepts hosted gateway outcome callbacks

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.

### Requirement: Payment reconciles Checkout-owned session expiry

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.

### Requirement: Payment snapshots commerce facts when creating a hosted session

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.

### Requirement: Hosted payment session is one-shot per order

**Reason**: The marketplace no longer runs a Payment service; the external payment gateway owns payment state and Checkout integrates with it directly.

**Migration**: See `payment-gateway-integration` and the modified `checkout` requirements.
