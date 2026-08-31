## Purpose

Define strict Cilium mutual TLS, identity-based authorization, retries, and circuit breakers for marketplace traffic after Istio has been removed.

## ADDED Requirements

### Requirement: Strict mTLS for marketplace east-west

Marketplace service-to-service traffic in `ecommerce` SHALL require Cilium mutual authentication (mTLS). Workloads that are not in the documented exception list (CNPG, migration jobs, Valkey sidecars, Kafka brokers in `kafka`) SHALL fail closed when mTLS is missing.

#### Scenario: Enrolled gRPC calls require mTLS

- **WHEN** `web` calls an internal gRPC Service after enforcement is enabled
- **THEN** the connection is mutually authenticated with Cilium identities

#### Scenario: Kafka TLS is not intercepted as mesh mTLS

- **WHEN** marketplace pods speak to Strimzi brokers in the `kafka` namespace
- **THEN** those connections are not required to use Cilium mesh mTLS in place of Kafka TLS

### Requirement: Authorization between marketplace identities

The system SHALL allow only documented caller identities to reach each marketplace HTTP/gRPC Service (equivalent intent to Istio AuthorizationPolicy).

#### Scenario: Unknown identity is denied

- **WHEN** a pod that is not an allowed caller sends traffic to a protected gRPC Service
- **THEN** Hubble or Cilium policy drops the flow and the call fails

#### Scenario: Documented callers succeed

- **WHEN** `web` calls `orders` (or other documented pairs) after policies are enforced
- **THEN** checkout and related flows continue to succeed

### Requirement: Retries and circuit breakers

The system SHALL apply documented retry and circuit-breaker (or outlier-detection) settings for Gateway-to-web and/or web-to-gRPC paths using Cilium/Envoy or Gateway API, without retrying non-idempotent checkout/payment RPCs unsafely.

#### Scenario: Transient failures retry where safe

- **WHEN** a safe idempotent hop fails with a retryable error
- **THEN** the dataplane retries within the documented budget

#### Scenario: Unhealthy backends are ejected

- **WHEN** a backend instance exceeds the circuit-breaker / outlier threshold
- **THEN** new requests are not sent to that instance until it recovers or the canary is aborted

### Requirement: Policy rollback

The system SHALL document how to disable strict mTLS and authorization without application code changes so the shop can be restored if enforcement breaks traffic.

#### Scenario: Enforcement disabled

- **WHEN** mesh policy enforcement is rolled back per docs
- **THEN** marketplace Services accept traffic as they did after the Cilium ingress cutover (no Istio)
