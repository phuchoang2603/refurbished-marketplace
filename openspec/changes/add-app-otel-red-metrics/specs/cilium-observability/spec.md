## MODIFIED Requirements

### Requirement: Hubble is not a closure requirement

After Istio removal, marketplace network observability SHALL NOT require Hubble (relay/UI), Hubble HTTP/gRPC metrics, or CiliumNetworkPolicy L7 visibility rules. Hubble MAY remain disabled or deleted on the cluster. Application request/error/duration SLIs SHALL come from marketplace OpenTelemetry metrics in VictoriaMetrics, not from Hubble or Istio.

#### Scenario: L7 mesh metrics are not required

- **WHEN** marketplace flows are verified after the Cilium cutover
- **THEN** closure does not depend on `hubble_http_*` metrics, Hubble UI, or Istio `istio_requests_total`

#### Scenario: App RED uses OTEL metrics

- **WHEN** contributors look for request/error/duration SLIs after Istio RED is removed
- **THEN** they use the Marketplace RED dashboard backed by application OpenTelemetry metrics rather than restoring Hubble scrapes

### Requirement: Istio mesh telemetry is absent

The observability chart SHALL NOT scrape Istio waypoint or Istio ingress proxies. The Marketplace Istio RED dashboard SHALL NOT be deployed. A Grafana dashboard that uses application OpenTelemetry metrics MAY be deployed.

#### Scenario: Istio scrapes disabled

- **WHEN** the observability chart is rendered after this change
- **THEN** it does not create VMPodScrapes for `ecommerce-waypoint` or Istio ingress Envoy stats

#### Scenario: Istio RED dashboard removed

- **WHEN** Grafana marketplace dashboards are synced
- **THEN** Marketplace Istio RED is not present
