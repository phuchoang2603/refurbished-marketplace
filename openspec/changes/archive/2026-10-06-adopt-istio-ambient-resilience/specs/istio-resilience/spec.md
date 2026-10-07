## Purpose

Use Istio's native ingress and ambient waypoint controls to keep marketplace requests responsive under transient failures without replaying unsafe buyer mutations or replacing the durable checkout saga.

## ADDED Requirements

### Requirement: HTTP and gRPC traffic has bounded mesh protection

Selected marketplace HTTP and gRPC Services SHALL use an Istio waypoint for L7 traffic policy and service-scoped limits on concurrent connections and pending requests. Endpoint outlier ejection SHALL only be enabled when the destination has multiple healthy instances; Kafka and database protocols SHALL NOT receive HTTP/gRPC retry or circuit-breaking policy.

#### Scenario: Healthy alternatives exist

- **WHEN** one of multiple endpoints repeatedly fails
- **THEN** the proxy can stop routing to that endpoint while continuing to serve traffic through healthy instances

#### Scenario: Destination has one endpoint

- **WHEN** a protected Service has a single healthy endpoint
- **THEN** resilience policy does not eject its only endpoint and turn an otherwise available Service into an outage

### Requirement: Retries preserve business identity

Mesh-wide default HTTP/gRPC retries SHALL be disabled, and bounded retries SHALL be explicitly configured only for operations whose inputs and effects can be replayed safely. Replayed `SubmitCheckout` requests SHALL retain the same buyer, intent key, and exact payload and SHALL return the original checkout instead of creating another order or payment attempt. Read-only requests MAY use bounded transient-failure retries. Retry attempts and per-try timeouts MUST fit within the enclosing request deadline.

#### Scenario: Checkout acknowledgement is lost

- **WHEN** an identical buyer-authenticated `SubmitCheckout` unary RPC is retried after Checkout committed but its response was lost
- **THEN** it returns the same checkout identifier and does not issue a second effective order command

#### Scenario: Retry changes checkout facts

- **WHEN** a repeat intent contains changed cart, price, shipping, or other snapshot facts
- **THEN** it is rejected as a conflict rather than silently creating or replaying a different checkout

#### Scenario: Read-only call encounters a transient error

- **WHEN** a configured lookup encounters a retryable transport failure within its deadline
- **THEN** the mesh MAY retry the same request a bounded number of times and return its successful result

### Requirement: Mutating browser requests are not blindly replayed

Ingress SHALL NOT automatically retry browser checkout POSTs, cart increments, hosted-payment submissions, or payment callback POSTs. Mesh retry policies SHALL NOT retry payment or cart mutations solely because their Service is reachable through the waypoint. Duplicate provider callbacks SHALL remain safe at Payment without repeating browser cart side effects.

#### Scenario: Checkout POST response is lost

- **WHEN** Web's checkout POST may have been processed but the browser did not receive a redirect
- **THEN** ingress does not rerun Web's cart and product snapshot assembly automatically

#### Scenario: Gateway callback is redelivered

- **WHEN** the payment provider sends the same terminal callback again
- **THEN** Payment preserves one definitive outcome and Web does not prematurely drain or duplicate cart changes

### Requirement: Route timeouts and retry budgets are visible

Shop and pay HTTPRoutes SHALL bound total and per-backend request time. The configured mesh resilience policies SHALL cap retry pressure and expose timeout, retry, rejection, and outlier signals to the existing platform telemetry pipeline without replacing application OTLP spans or metrics.

#### Scenario: Backend stalls

- **WHEN** an HTTP backend exceeds its time budget
- **THEN** the request terminates within the documented total deadline rather than retrying indefinitely

#### Scenario: Transient failure is observed

- **WHEN** a waypoint retries or limits a backend request
- **THEN** operators can distinguish mesh retry, rejection, and endpoint health signals from application-level checkout progress

### Requirement: The saga owns durable recovery

Istio resilience controls SHALL NOT replace Checkout, Orders, Inventory, or Payment inbox/outbox replay, durable deadlines, compensation, and financial-exception handling. Istio SHALL NOT retry Kafka message delivery in lieu of application-level processing.

#### Scenario: Buyer disconnects after acceptance

- **WHEN** the browser disconnects after Checkout commits an intent
- **THEN** the saga continues or reconciles through its existing durable mechanisms independently of any proxy retry
