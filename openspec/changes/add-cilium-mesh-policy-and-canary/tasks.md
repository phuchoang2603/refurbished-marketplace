## 1. Prerequisites

- [ ] 1.1 Confirm `replace-istio-with-cilium` is archived and shop/pay traffic uses Cilium Gateway.
- [ ] 1.2 Write the allowed-caller matrix (web → gRPC services, payment ↔ simulator, jobs/DBs/Valkey exceptions, kafka left on TLS).

## 2. mTLS and authorization

- [ ] 2.1 Add Cilium mutual-auth / SPIFFE settings required for `ecommerce` identities.
- [ ] 2.2 Add CiliumNetworkPolicies (observe then enforce) for documented caller pairs; deny unknown identities.
- [ ] 2.3 Document Hubble verification of allowed vs dropped flows and rollback to non-enforcing.

## 3. Retries and circuit breakers

- [ ] 3.1 Configure retry/timeout/outlier settings via Gateway API and/or CiliumEnvoyConfig for safe hops only.
- [ ] 3.2 Confirm CreateOrder / hosted-payment callback are not retried unsafely; align with existing idempotency.

## 4. Argo canary

- [ ] 4.1 Add Argo Rollouts wrapper chart and app-of-apps Application.
- [ ] 4.2 Convert `web` (or chosen first service) to a Rollout with Gateway API HTTPRoute weights on the Cilium Gateway.
- [ ] 4.3 Document Tilt vs staging: how local still deploys, and how to abort a canary.

## 5. Verify and close

- [ ] 5.1 Staging: enforce mTLS/authz, complete checkout, abort a canary, restore stable.
- [ ] 5.2 Run `openspec validate add-cilium-mesh-policy-and-canary` and update GitHub issue [#39](https://github.com/phuchoang2603/refurbished-marketplace/issues/39).
