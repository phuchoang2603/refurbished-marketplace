## Why

Issue [#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39): After Istio removal, east–west traffic is still plain cluster networking: no mandatory mTLS, no identity-based authorization, and no Gateway timeout guardrails. This is the deferred hardening slice from the original Istio observe baseline, implemented on Cilium.

Depends on archived `replace-istio-with-cilium` / [#38](https://github.com/phuchoang2603/refurbished-marketplace/issues/38).

## What Changes

- Enforce Cilium mutual auth on enrolled marketplace hops (`CiliumNetworkPolicy` `authentication.mode: required` + `enableDefaultDeny.ingress: true` when enforced). SPIRE stays in talos-proxmox.
- Add identity allow-lists for documented callers. Kafka/Strimzi TLS stays outside mesh mTLS.
- Add HTTPRoute `timeouts` on shop/pay. No HTTPRoute retries or gRPC retry interceptors (checkout idempotency is #35).
- Rely on Cilium Gateway Envoy outlier detection for backend ejection; no `CiliumEnvoyConfig` in this repo.
- Document rollback: disable strict auth without taking the shop down.

## Capabilities

### New Capabilities

- `cilium-mesh-policy`: Strict mTLS, authorization, Gateway timeouts, and documented outlier-detection posture for marketplace traffic on Cilium.

### Modified Capabilities

- (none)

## Impact

- Marketplace chart: `CiliumNetworkPolicy` objects and HTTPRoute timeouts.
- talos-proxmox: SPIRE / mutual-auth Helm must be enabled before `meshPolicy.mutualAuth: true`.
- Out of scope: canary, Argo Rollouts, HTTP/gRPC retries on checkout until #35.
