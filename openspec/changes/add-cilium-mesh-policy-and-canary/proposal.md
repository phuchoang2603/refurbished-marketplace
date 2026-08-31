## Why

Issue [#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39): After Istio is gone, east–west traffic is still plain cluster networking: no mandatory mTLS, no identity-based authorization, no mesh retries/circuit breakers, and deployments are all-at-once. This change is the deferred hardening + progressive-delivery slice that the original Istio observe baseline explicitly left out.

Depends on `replace-istio-with-cilium` / [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38) landing first (Cilium Gateway API edge, no Istio).

## What Changes

- Enforce strict mutual TLS on marketplace east–west traffic using Cilium (SPIFFE / Cilium mutual auth), not Istio `PeerAuthentication`.
- Add authorization for marketplace identities (CiliumNetworkPolicy and/or equivalent) so only intended clients reach gRPC/HTTP services. Kafka/Strimzi TLS stays outside L7 intercept.
- Add retries and circuit-breaking for east–west and/or Gateway north–south using Cilium/Envoy (or Gateway API retry/timeout) rather than Istio `VirtualService` / `DestinationRule`.
- Add Argo Rollouts (or equivalent Argo progressive delivery) so at least one marketplace workload can canary via Gateway API HTTPRoute weights (or Cilium GAMMA) instead of a single ReplicaSet cutover.
- Document rollback: disable strict auth without taking the shop down; abort a canary back to stable.

## Capabilities

### New Capabilities

- `cilium-mesh-policy`: Strict mTLS, authorization, retries, and circuit breakers for marketplace traffic on Cilium after the Istio removal.
- `argocd-canary`: GitOps-managed canary (Argo Rollouts + Gateway API traffic splitting) for marketplace workloads.

### Modified Capabilities

- `argocd-gitops`: Add a Rollouts (or equivalent) operator Application and canary-aware marketplace delivery. Do not resurrect Istio. Canary traffic splitting attaches to the Cilium Gateway HTTPRoutes introduced by `replace-istio-with-cilium`.

## Impact

- New platform operator (Argo Rollouts) plus Cilium policy objects and possibly `CiliumEnvoyConfig` / HTTPRoute filters.
- Marketplace chart Deployments may become Rollouts for chosen services (likely `web` first).
- Observability should show canary vs stable and mTLS/authz denials (Hubble + existing Grafana).
- Out of scope until `replace-istio-with-cilium` is archived: implementing any of this against Istio CRDs.
