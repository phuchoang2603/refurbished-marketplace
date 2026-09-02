## 1. Prerequisites

- [x] 1.1 Confirm Cilium Gateway is the live shop/pay origin on Talos.
- [x] 1.2 Confirm talos-proxmox Cilium Helm has mutual-auth/SPIRE before enforcing `authentication.mode: required`.
- [x] 1.3 Write the allowed-caller matrix (web → gRPC services, payment ↔ simulator, jobs/DBs/Valkey exceptions, kafka left on TLS).

## 2. mTLS and authorization

- [x] 2.1 Add CiliumNetworkPolicies with `enableDefaultDeny.ingress: true` when enforced; deny unknown identities.
- [x] 2.2 Set `authentication.mode: required` on enrolled hops.
- [x] 2.3 Document policy verification (cilium verdicts / agent logs / traces) and rollback to non-enforcing.

## 3. Gateway timeouts

- [x] 3.1 Configure HTTPRoute `timeouts` on shop and pay; no HTTPRoute retries or `CiliumEnvoyConfig`.
- [x] 3.2 Confirm CreateOrder / hosted-payment callback are not retried at the dataplane (gRPC clients have no retry interceptor).

## 4. Verify and close

- [x] 4.1 talos-dev: enforce mTLS/authz, complete checkout, confirm a probe pod is denied.
- [x] 4.2 Run `openspec validate add-cilium-mesh-policy-and-canary` and update GitHub issue [#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39).
