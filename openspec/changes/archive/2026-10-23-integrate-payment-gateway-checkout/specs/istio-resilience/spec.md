## MODIFIED Requirements

### Requirement: Mutating browser requests are not blindly replayed

Ingress SHALL NOT automatically retry browser checkout POSTs, cart increments, or gateway webhook POSTs. Mesh retry policies SHALL NOT retry checkout payment or cart mutations solely because their Service is reachable through the waypoint. Duplicate gateway webhooks SHALL remain safe at Checkout without repeating browser cart side effects.

#### Scenario: Checkout POST response is lost

- **WHEN** Web's checkout POST may have been processed but the browser did not receive a redirect
- **THEN** ingress does not rerun Web's cart and product snapshot assembly automatically

#### Scenario: Gateway callback is redelivered

- **WHEN** the payment gateway delivers the same event ID again
- **THEN** Checkout preserves one definitive outcome and Web does not prematurely drain or duplicate cart changes

### Requirement: Route timeouts and retry budgets are visible

Shop HTTPRoutes SHALL bound total and per-backend request time. The configured mesh resilience policies SHALL cap retry pressure and expose timeout, retry, rejection, and outlier signals to the existing platform telemetry pipeline without replacing application OTLP spans or metrics.

#### Scenario: Backend stalls

- **WHEN** an HTTP backend exceeds its time budget
- **THEN** the request terminates within the documented total deadline rather than retrying indefinitely

#### Scenario: Transient failure is observed

- **WHEN** a waypoint retries or limits a backend request
- **THEN** operators can distinguish mesh retry, rejection, and endpoint health signals from application-level checkout progress
