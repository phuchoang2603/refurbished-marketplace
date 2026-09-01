## 1. Prerequisites

- [ ] 1.1 Confirm Cilium Gateway is the live shop/pay origin on Talos.
- [ ] 1.2 Confirm talos-proxmox Cilium Helm has mutual-auth/SPIRE before enforcing `authentication.mode: required`.
- [ ] 1.3 Write the allowed-caller matrix (web → gRPC services, payment ↔ simulator, jobs/DBs/Valkey exceptions, kafka left on TLS).

## 2. mTLS and authorization

- [ ] 2.1 Add CiliumNetworkPolicies (observe then enforce) for documented caller pairs; deny unknown identities.
- [ ] 2.2 Set `authentication.mode: required` on enrolled hops.
- [ ] 2.3 Document policy verification (cilium verdicts / agent logs / traces) and rollback to non-enforcing.

## 3. Retries and circuit breakers

- [ ] 3.1 Configure retry/timeout/outlier settings via Gateway API and/or CiliumEnvoyConfig for safe hops only.
- [ ] 3.2 Confirm CreateOrder / hosted-payment callback are not retried unsafely; align with existing idempotency.

## 4. Verify and close

- [ ] 4.1 talos-dev: enforce mTLS/authz, complete checkout, confirm a probe pod is denied.
- [ ] 4.2 Run `openspec validate add-cilium-mesh-policy-and-canary` and update GitHub issue [#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39).
