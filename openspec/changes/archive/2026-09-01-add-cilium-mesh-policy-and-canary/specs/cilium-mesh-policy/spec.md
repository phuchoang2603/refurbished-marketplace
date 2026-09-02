## Purpose

Define strict Cilium mutual TLS, identity-based authorization, Gateway timeouts, and Cilium Gateway outlier detection for marketplace traffic after Istio has been removed.

## ADDED Requirements

### Requirement: Strict mTLS for marketplace east-west

Enrolled marketplace hops in `ecommerce` SHALL use Cilium mutual authentication (`authentication.mode: required` on documented ingress rules). Workloads outside the exception list (CNPG, migration jobs, Valkey, Kafka brokers in `kafka`) are not selected by these policies.

#### Scenario: Enrolled gRPC calls require mTLS

- **WHEN** `web` calls an internal gRPC Service after enforcement is enabled
- **THEN** the connection is mutually authenticated with Cilium identities

#### Scenario: Kafka TLS is not intercepted as mesh mTLS

- **WHEN** marketplace pods speak to Strimzi brokers in the `kafka` namespace
- **THEN** those connections are not required to use Cilium mesh mTLS in place of Kafka TLS

### Requirement: Authorization between marketplace identities

The system SHALL allow only documented caller identities to reach each marketplace HTTP/gRPC Service. When `meshPolicy.enforce` is true, CNPs SHALL set `enableDefaultDeny.ingress: true` so unknown callers are dropped.

#### Scenario: Unknown identity is denied

- **WHEN** a pod that is not an allowed caller sends traffic to a protected gRPC Service
- **THEN** Cilium policy drops the flow and the call fails

#### Scenario: Documented callers succeed

- **WHEN** `web` calls `orders` (or other documented pairs) after policies are enforced
- **THEN** checkout and related flows continue to succeed

### Requirement: Gateway timeouts and outlier detection

The system SHALL set documented HTTPRoute `timeouts` on shop and pay routes. The system SHALL NOT configure HTTPRoute retries or gRPC client retries for checkout or payment paths. Circuit breaking for Gateway backends SHALL rely on Cilium Gateway Envoy outlier detection (no `CiliumEnvoyConfig` in this repo).

#### Scenario: Shop routes have request timeouts

- **WHEN** shop or pay HTTPRoutes are rendered with ingress enabled
- **THEN** each rule documents `timeouts.request` and `timeouts.backendRequest`

#### Scenario: Checkout is not retried at the dataplane

- **WHEN** a browser POST hits checkout or a hosted-payment callback on the shop hostname
- **THEN** the HTTPRoute does not configure retries and gRPC clients do not add retry interceptors

### Requirement: Policy rollback

The system SHALL document how to disable strict mTLS and authorization without application code changes so the shop can be restored if enforcement breaks traffic.

#### Scenario: Enforcement disabled

- **WHEN** mesh policy enforcement is rolled back per docs
- **THEN** marketplace Services accept traffic as they did after the Cilium ingress cutover (no Istio)
