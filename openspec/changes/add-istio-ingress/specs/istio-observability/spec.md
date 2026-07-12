## MODIFIED Requirements

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
