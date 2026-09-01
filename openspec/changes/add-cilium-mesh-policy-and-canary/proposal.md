## Why

Issue [#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39): After Istio removal, east–west traffic is still plain cluster networking: no mandatory mTLS, no identity-based authorization, and no mesh retries/circuit breakers. This is the deferred hardening slice from the original Istio observe baseline, implemented on Cilium.

Depends on archived `replace-istio-with-cilium` / [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38) (Cilium Gateway edge, no Istio).

## What Changes

- Enforce strict mutual TLS on marketplace east–west traffic using Cilium identities (`CiliumNetworkPolicy` `authentication.mode: required`), not Istio `PeerAuthentication`. SPIRE / Cilium mutual-auth Helm is enabled in talos-proxmox; this repo consumes it and does not helm-upgrade Cilium.
- Add authorization so only documented callers reach marketplace HTTP/gRPC Services. Kafka/Strimzi TLS stays outside L7 intercept.
- Add retries and circuit-breaking for Gateway → web and/or web → gRPC using Gateway API and/or `CiliumEnvoyConfig`, not Istio `VirtualService` / `DestinationRule`.
- Document rollback: disable strict auth without taking the shop down.

## Capabilities

### New Capabilities

- `cilium-mesh-policy`: Strict mTLS, authorization, retries, and circuit breakers for marketplace traffic on Cilium after Istio removal.

### Modified Capabilities

- (none)

## Impact

- Marketplace chart: CiliumNetworkPolicies and possibly `CiliumEnvoyConfig` / HTTPRoute retry filters.
- talos-proxmox: Cilium Helm enables mutual auth/SPIRE so `authentication.mode: required` can fail closed.
- Observability: policy drops should be visible with whatever the cluster already exposes (Hubble is optional/off on Talos today); do not re-enable Hubble as a hard dependency.
- Out of scope: Istio CRDs; Argo Rollouts / Flagger; canary or HTTPRoute traffic splitting; canary Deployments.
