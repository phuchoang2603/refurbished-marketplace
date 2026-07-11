## 1. Istio Platform Setup

- [ ] 1.1 Create a separate prerequisite proposal for VictoriaMetrics/VictoriaTraces with Grafana.
- [ ] 1.2 Confirm staging cluster Kubernetes version supports Istio ambient mode.
- [ ] 1.3 Confirm staging cluster networking/CNI compatibility for Istio ambient mode.
- [ ] 1.4 Confirm Gateway API CRDs are installed or planned for waypoint proxy support.
- [ ] 1.5 Add the official Istio Helm repository source for GitOps-managed installation.
- [ ] 1.6 Pin Istio platform charts to version `1.30.2`.
- [ ] 1.7 Add staging ArgoCD Application resources for Istio base/control plane/CNI/ztunnel components as required by ambient mode.
- [ ] 1.8 Set Istio sync ordering so platform resources apply before mesh-enrolled marketplace workloads.
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

- [ ] 4.1 Document VictoriaTraces with Grafana as the target trace and dashboard path.
- [ ] 4.2 Verify staging sync installs Istio and enrolls marketplace workloads.
- [ ] 4.3 Verify Istio CNI and ztunnel pods are healthy in `istio-system`.
- [ ] 4.4 Exercise product, cart, checkout, and payment flows in staging.
- [ ] 4.5 Confirm mesh telemetry shows traffic for `web`, `users`, `products`, `orders`, `cart`, `payment`, and `payment-gateway-simulator` where applicable.
- [ ] 4.6 Confirm internal gRPC traffic is distinguishable from opaque TCP where Istio supports protocol detection.
- [ ] 4.7 Keep Grafana/VictoriaTraces dashboard checks blocked on the separate prerequisite observability change.

## 5. Final Checks

- [ ] 5.1 Run OpenSpec validation for `add-istio-observe-baseline`.
- [ ] 5.2 Update GitHub issue #18 with ambient mode direction, waypoint proxy plan, official Istio `1.30.2` pin, telemetry prerequisite, and staging-only production gate.
