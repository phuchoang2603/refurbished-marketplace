## 1. Istio Platform Setup

- [x] 1.1 Confirm VictoriaMetrics/VictoriaTraces with Grafana prerequisite (`platform-observability`) is implemented and archived.
- [x] 1.2 Confirm staging cluster Kubernetes version supports Istio ambient mode (`v1.32.3+rke2r1`).
- [x] 1.3 Confirm staging cluster networking/CNI compatibility for Istio ambient mode (RKE2 Canal; allow HBONE TCP 15008 in NetworkPolicies).
- [x] 1.4 Confirm Gateway API CRDs are installed (present on staging; GatewayClass/controller deferred until waypoints).
- [ ] 1.5 Add four local wrapper charts under `infra/charts/operators/istio/{base,istiod,cni,ztunnel}` sourcing the official Istio Helm charts.
- [ ] 1.6 Pin each Istio wrapper chart dependency to version `1.30.2` (`istiod`/`cni` with `profile=ambient`).
- [ ] 1.7 Add four staging ArgoCD Applications for the Istio wrappers (`base`, `istiod`, `cni`, `ztunnel`) in `istio-system`.
- [ ] 1.8 Set Istio sync waves so base → istiod/cni → ztunnel apply before mesh-enrolled marketplace workloads.
- [ ] 1.9 Keep production Istio installation and enrollment disabled until staging is verified.

## 2. Marketplace Mesh Enrollment

- [ ] 2.1 Add staging marketplace ambient enrollment through GitOps-managed namespace or workload metadata.
- [ ] 2.2 Add waypoint proxy configuration for workloads that need L7 telemetry, policy, or routing behavior.
- [ ] 2.3 Add rollback notes for disabling marketplace mesh enrollment before removing Istio.

## 3. Protocol-Aware Service Ports

- [ ] 3.1 Add per-service protocol configuration to the marketplace Helm values.
- [ ] 3.2 Render HTTP port names for `web` and `payment-gateway-simulator`.
- [ ] 3.3 Render gRPC port names for `users`, `products`, `orders`, `cart`, and `payment`.
- [ ] 3.4 Verify Helm output no longer labels gRPC service ports as generic `http`.

## 4. Observability Verification

- [x] 4.1 Document VictoriaTraces with Grafana as the target trace and dashboard path (`docs/observability.md`).
- [ ] 4.2 Verify staging sync installs Istio and enrolls marketplace workloads.
- [ ] 4.3 Verify Istio CNI and ztunnel pods are healthy in `istio-system`.
- [ ] 4.4 Exercise product, cart, checkout, and payment flows in staging.
- [ ] 4.5 Confirm mesh telemetry shows traffic for `web`, `users`, `products`, `orders`, `cart`, `payment`, and `payment-gateway-simulator` where applicable.
- [ ] 4.6 Confirm internal gRPC traffic is distinguishable from opaque TCP where Istio supports protocol detection.
- [ ] 4.7 Verify Istio metrics/traces/dashboards against the deployed `platform-observability` stack in Grafana / VictoriaTraces.

## 5. Final Checks

- [ ] 5.1 Run OpenSpec validation for `add-istio-observe-baseline`.
- [ ] 5.2 Update GitHub issue #18 with ambient mode direction, waypoint proxy plan, official Istio `1.30.2` pin, telemetry prerequisite, and staging-only production gate.
