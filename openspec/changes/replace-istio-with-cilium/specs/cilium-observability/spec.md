## Purpose

Define post-Istio observe-only networking telemetry: Hubble L4 flows plus existing application OpenTelemetry traces, without Cilium L7 visibility policies or Istio waypoint metrics.

## ADDED Requirements

### Requirement: Hubble L4 is the network observe path

After Istio removal, marketplace network observability SHALL use Hubble L4 (flows / Hubble UI) and SHALL NOT require Hubble HTTP/gRPC metrics, CiliumNetworkPolicy L7 visibility rules, or a Grafana dashboard that replaces Marketplace Istio RED.

#### Scenario: L7 mesh metrics are not required

- **WHEN** marketplace flows are verified after the Cilium cutover
- **THEN** closure does not depend on `hubble_http_*` metrics or Istio `istio_requests_total`

#### Scenario: Hubble is available where Cilium runs

- **WHEN** Cilium is installed for local or staging per documented values
- **THEN** Hubble (including relay/UI when those flags are enabled) can show L4 flows involving marketplace pods

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

- **WHEN** the marketplace namespace is applied by Argo CD or Tilt
- **THEN** it does not set `istio.io/dataplane-mode` or `istio.io/use-waypoint`
