## REMOVED Requirements

### Requirement: Protocol-aware service ports

**Reason**: Protocol metadata remains necessary but is no longer a Cilium observability concern.
**Migration**: None; the new ambient networking contract retains protocol-aware ports.

#### Scenario: Service ports

- **WHEN** application Services render
- **THEN** their gRPC/HTTP names are not tied to Cilium observability

### Requirement: No observe-only Istio enrollment

**Reason**: Marketplace and Kafka namespaces explicitly enroll in Istio ambient.
**Migration**: None; use the new ambient networking contract.

#### Scenario: New namespace

- **WHEN** application namespaces are created
- **THEN** no prohibition on Istio ambient labels remains

### Requirement: Application traces use OTEL export

**Reason**: Application OTLP tracing remains in its existing tracing and metrics capabilities, not a retired Cilium-specific one.
**Migration**: None; preserve OTLP export to the platform agent.

#### Scenario: Trace export

- **WHEN** Web handles checkout
- **THEN** application spans do not require a Cilium observability contract

### Requirement: Hubble is not required for observability

**Reason**: Istio dataplane signals and application OTLP metrics supersede the old post-Istio restrictions.
**Migration**: None; use the new resilience observability contract.

#### Scenario: Mesh failure diagnosis

- **WHEN** an operator investigates retry and ejection behavior
- **THEN** the old restriction on Istio metrics does not apply

### Requirement: Marketplace renders no Istio telemetry

**Reason**: Istio waypoint and ingress telemetry are needed to verify resilience; the platform owns collection.
**Migration**: None; use the platform telemetry integration in `istio-resilience`.

#### Scenario: Resilience observation

- **WHEN** a proxy retries or rejects a request
- **THEN** its signal may be exported through the platform pipeline

### Requirement: No inbound telemetry exception

**Reason**: This CiliumNetworkPolicy-only requirement cannot apply once those policies are removed.
**Migration**: None; application OTLP export remains unchanged.

#### Scenario: Mesh policy render

- **WHEN** Helm renders the marketplace release
- **THEN** no Cilium telemetry allow-list is required
