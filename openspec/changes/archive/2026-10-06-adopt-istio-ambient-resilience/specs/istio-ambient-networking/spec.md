## Purpose

Provide marketplace browser ingress and authenticated east-west transport using Istio ambient on Talos while keeping Cilium as the platform networking provider.

## ADDED Requirements

### Requirement: Marketplace namespaces use ambient transport

Each environment SHALL enroll `ecommerce` and `kafka` declaratively in Istio ambient and SHALL require mutual TLS for mesh-to-mesh workload traffic in those namespaces. Cilium SHALL remain the platform-owned CNI and load-balancing provider; the marketplace SHALL NOT install Cilium or render CiliumNetworkPolicy or Cilium mutual-authentication resources.

#### Scenario: Marketplace calls traverse ambient

- **WHEN** `web` calls a marketplace gRPC Service or Kafka Connect calls MongoDB across the two enrolled namespaces
- **THEN** the connections use Istio mutual TLS, and no Cilium-specific authentication policy is required

#### Scenario: Cluster-owned CNI remains available

- **WHEN** the marketplace root syncs against a newly provisioned environment
- **THEN** it does not install or upgrade Cilium and the application uses the existing pod and Service networking

### Requirement: Namespace enrollment is reconciled by GitOps

The marketplace GitOps root SHALL ensure both namespace labels are present before relying on strict mesh traffic; shared operators and platform namespaces SHALL NOT be enrolled by the marketplace. Workload health and operator reconciliation MUST continue to work with strict mTLS.

#### Scenario: Empty cluster convergence

- **WHEN** the production marketplace root syncs onto a fresh platform with no `ecommerce` or `kafka` namespace
- **THEN** both namespaces acquire the ambient enrollment label without an out-of-band labeling command

#### Scenario: Operator and probe access

- **WHEN** marketplace databases, brokers, connectors, and application pods are reconciled
- **THEN** their management paths and health probes continue to function without bypassing required workload mutual TLS

### Requirement: Marketplace uses Istio Gateway API ingress

When ingress is enabled, the marketplace SHALL render a programmed `istio` Gateway and HTTPRoutes for web and the hosted-payment simulator. Cloudflare Tunnel SHALL remain the public HTTPS front door; its in-cluster HTTP origin SHALL resolve the Istio-managed Gateway Service without a LAN VIP or an application TLS certificate. Dev hosts SHALL be `shop-dev` / `pay-dev` and production hosts SHALL be `shop` / `pay` through the existing overlay.

#### Scenario: Web and simulator host routing

- **WHEN** a browser uses either configured hostname
- **THEN** the Istio Gateway routes that request to the corresponding Service, and unknown hosts have no marketplace route

#### Scenario: Cloudflare origin is reconciled

- **WHEN** the tunnel binding and Gateway converge
- **THEN** both hostnames resolve to the in-cluster Istio Gateway origin and forwarded HTTPS host/scheme semantics are preserved

### Requirement: Protocol metadata remains accurate

Marketplace Services SHALL retain HTTP and gRPC port metadata compatible with Istio traffic classification. Application OTLP traces and RED metrics SHALL continue to reach the platform agent independently of mesh proxy metrics.

#### Scenario: Services are classified

- **WHEN** application Services are rendered
- **THEN** their ports identify HTTP and gRPC traffic by the actual application protocol

#### Scenario: Application telemetry is retained

- **WHEN** the ingress and ambient dataplanes replace Cilium mesh behavior
- **THEN** existing application traces and request metrics continue to flow through `otel-agent`
