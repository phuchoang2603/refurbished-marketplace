# Cilium Observability

## Purpose

Define post-Istio observe path: marketplace OpenTelemetry traces to VictoriaTraces, without Hubble, Cilium L7 visibility policies, or Istio waypoint metrics. Application-level RED metrics are OpenTelemetry metrics in VictoriaMetrics (not Hubble).

## Requirements

### Requirement: Hubble is not a closure requirement

After Istio removal, marketplace network observability SHALL NOT require Hubble (relay/UI), Hubble HTTP/gRPC metrics, CiliumNetworkPolicy L7 visibility rules, or a Grafana dashboard that replaces Marketplace Istio RED. Hubble MAY be disabled or deleted on the cluster.

#### Scenario: L7 mesh metrics are not required

- **WHEN** marketplace flows are verified after the Cilium cutover
- **THEN** closure does not depend on `hubble_http_*` metrics, Hubble UI, or Istio `istio_requests_total`

#### Scenario: App RED is a follow-on

- **WHEN** contributors look for request/error/duration SLIs after Istio RED is removed
- **THEN** documentation points at GitHub issue #43 (app-level OTEL metrics) rather than restoring Hubble scrapes

### Requirement: Application traces remain OTEL to VictoriaTraces

Distributed tracing SHALL continue to use marketplace OpenTelemetry export to VictoriaTraces. Cilium Gateway and Hubble SHALL NOT be required to emit proxy spans for checkout verification.

#### Scenario: Waterfall is application spans

- **WHEN** a contributor inspects a checkout trace in Grafana
- **THEN** verification uses service OTEL spans (and Kafka/outbox continuation) rather than Gateway or Hubble spans

### Requirement: Istio mesh telemetry is absent

The observability chart SHALL NOT scrape Istio waypoint or Istio ingress proxies. The Marketplace Istio RED dashboard SHALL NOT be deployed.

#### Scenario: Istio scrapes disabled

- **WHEN** the observability chart is rendered after this change
- **THEN** it does not create VMPodScrapes for `ecommerce-waypoint` or Istio ingress Envoy stats

#### Scenario: Istio RED dashboard removed

- **WHEN** Grafana marketplace dashboards are synced
- **THEN** Marketplace Istio RED is not present

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
