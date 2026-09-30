# Cilium Observability

## Purpose

Define post-Istio observe path: marketplace OpenTelemetry traces exported to the platform `otel-agent`, without Hubble, Cilium L7 visibility policies, or Istio waypoint metrics. Application-level RED metrics are OpenTelemetry metrics pushed over OTLP to the same agent (not Hubble).

## Requirements

### Requirement: Protocol-aware service ports

The system SHALL expose Kubernetes Service ports with names (and appProtocol where rendered) that match the protocol used by each marketplace service.

#### Scenario: gRPC service ports are named as gRPC

- **WHEN** the marketplace Helm chart renders Services for internal gRPC services
- **THEN** the rendered Service port names identify the ports as gRPC rather than generic HTTP

#### Scenario: HTTP service ports remain HTTP

- **WHEN** the marketplace Helm chart renders Services for browser-facing or HTTP-only workloads
- **THEN** the rendered Service port names identify the ports as HTTP

### Requirement: No observe-only Istio enrollment

Marketplace namespaces SHALL NOT be labeled for Istio ambient dataplane or waypoint. Workloads SHALL communicate over Kubernetes Services on Cilium without Istio ztunnel.

#### Scenario: Ambient labels absent

- **WHEN** the marketplace namespace is applied by Argo CD on Talos
- **THEN** it does not set `istio.io/dataplane-mode` or `istio.io/use-waypoint`

### Requirement: Application traces use OTEL export

Distributed tracing SHALL use marketplace OpenTelemetry export to the platform agent. Cilium Gateway and Hubble SHALL NOT be required to emit proxy spans for checkout verification.

#### Scenario: Waterfall is application spans

- **WHEN** a contributor inspects a checkout trace in HyperDX
- **THEN** verification uses service OTEL spans (and Kafka/outbox continuation) rather than Gateway or Hubble spans

### Requirement: Hubble is not required for observability

Marketplace network observability SHALL NOT require Hubble (relay/UI), Hubble HTTP/gRPC metrics, or CiliumNetworkPolicy L7 visibility rules. Hubble MAY remain disabled or deleted on the cluster. Application request/error/duration SLIs SHALL come from marketplace OpenTelemetry metrics pushed to the platform agent, not from Hubble or Istio.

#### Scenario: L7 mesh metrics are not required

- **WHEN** marketplace flows are verified
- **THEN** closure does not depend on `hubble_http_*` metrics, Hubble UI, or Istio `istio_requests_total`

#### Scenario: App RED uses pushed OTEL metrics

- **WHEN** contributors look for request/error/duration SLIs
- **THEN** they query application OpenTelemetry metrics in HyperDX rather than restoring Hubble scrapes

### Requirement: Marketplace renders no Istio telemetry

The marketplace release SHALL NOT scrape Istio waypoint or Istio ingress proxies, and SHALL NOT deploy an Istio RED dashboard.

#### Scenario: Istio scrapes absent

- **WHEN** the marketplace chart is rendered
- **THEN** it contains no scrape resources for `ecommerce-waypoint` or Istio ingress Envoy stats

#### Scenario: Istio RED dashboard absent

- **WHEN** the marketplace chart is rendered
- **THEN** it contains no Marketplace Istio RED dashboard

### Requirement: No inbound telemetry exception

Marketplace CiliumNetworkPolicies SHALL NOT allow inbound traffic for telemetry collection. Telemetry SHALL leave marketplace pods only as outbound OTLP to the platform agent, and marketplace policies SHALL NOT block that outbound traffic.

#### Scenario: No scrape policy is rendered

- **WHEN** the marketplace Helm chart renders mesh policy
- **THEN** it renders no `allow-metrics-scrape` policy and no policy that admits traffic from the `monitoring` or `observability` namespaces

#### Scenario: OTLP export is not blocked

- **WHEN** mesh policy is enforced with default-deny ingress
- **THEN** marketplace services still deliver OTLP to `otel-agent` in `observability`
