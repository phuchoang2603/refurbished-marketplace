# Istio Observability

## Purpose

Define observe-only Istio mesh enrollment for marketplace workloads, protocol-aware service telemetry, and rollback expectations without requiring application business logic changes or strict mesh policy enforcement.

## Requirements

### Requirement: Observe-only mesh enrollment

The system SHALL support enrolling marketplace workloads in Istio for telemetry collection without requiring application business logic changes or strict mesh policy enforcement.

#### Scenario: Marketplace workloads enroll in mesh

- **WHEN** the staging marketplace deployment is synced with mesh enrollment enabled
- **THEN** the marketplace workloads run as mesh participants and continue serving the existing web, product, cart, checkout, and payment flows

#### Scenario: Official Istio release is pinned

- **WHEN** Istio platform resources are configured for staging
- **THEN** the configuration uses local wrapper charts under `infra/charts/operators/istio/` that pin official Istio Helm chart versions to `1.30.2`

#### Scenario: Ambient prerequisites are verified before enrollment

- **WHEN** marketplace ambient enrollment is planned for staging
- **THEN** cluster Kubernetes version, Istio CNI readiness, ztunnel readiness, and Gateway API availability are verified before relying on ambient mode and waypoint proxies

#### Scenario: Ambient mode is preferred when prerequisites are met

- **WHEN** the staging cluster supports Istio ambient mode prerequisites
- **THEN** marketplace mesh enrollment uses ambient mode as the target direction

#### Scenario: Waypoint proxy is available for L7 behavior

- **WHEN** marketplace workloads require Istio L7 telemetry, policy, or routing behavior in ambient mode
- **THEN** the design provides a waypoint proxy path for those workloads

#### Scenario: Mesh enrollment remains non-disruptive

- **WHEN** workloads are enrolled for the observe-only baseline without ingress enablement
- **THEN** strict mTLS, AuthorizationPolicy, retries, circuit breakers, and traffic splitting are not required for the application to function

#### Scenario: Ingress is a separate explicit enablement

- **WHEN** marketplace Istio ingress is required for browser traffic
- **THEN** edge Gateway API resources are enabled through the dedicated ingress configuration path rather than being implied by observe-only mesh enrollment alone

### Requirement: Protocol-aware service classification

The system SHALL expose Kubernetes Service ports with names that match the protocol used by each marketplace service so Istio can classify traffic accurately.

#### Scenario: gRPC service ports are named as gRPC

- **WHEN** the marketplace Helm chart renders Services for internal gRPC services
- **THEN** the rendered Service port names identify the ports as gRPC rather than generic HTTP

#### Scenario: HTTP service ports remain HTTP

- **WHEN** the marketplace Helm chart renders Services for browser-facing or HTTP-only workloads
- **THEN** the rendered Service port names identify the ports as HTTP

### Requirement: Mesh telemetry visibility

The system SHALL provide observable Istio telemetry for marketplace service-to-service traffic in staging.

#### Scenario: Internal traffic appears in telemetry

- **WHEN** a user exercises the primary marketplace flows in staging
- **THEN** mesh telemetry shows traffic involving `web`, `users`, `products`, `orders`, `cart`, `payment`, and `payment-gateway-simulator` where applicable

#### Scenario: gRPC traffic is distinguishable

- **WHEN** the web service calls internal gRPC services during staging verification
- **THEN** telemetry distinguishes gRPC service calls from opaque TCP traffic where Istio protocol support allows it

#### Scenario: Grafana and VictoriaTraces are the target telemetry stack

- **WHEN** trace and dashboard verification is documented
- **THEN** the documentation targets VictoriaTraces with Grafana as the long-term telemetry visualization path

#### Scenario: Telemetry verification uses platform observability

- **WHEN** Istio trace and dashboard verification runs in staging
- **THEN** it uses the deployed `platform-observability` stack (`monitoring` namespace, Grafana / VictoriaTraces) rather than a temporary tracing UI

### Requirement: Mesh enrollment rollback

The system SHALL document and support rollback from marketplace mesh enrollment without requiring application code changes.

#### Scenario: Enrollment is disabled

- **WHEN** mesh enrollment is removed or disabled for marketplace workloads
- **THEN** the application returns to the previous Kubernetes Service-based communication path
