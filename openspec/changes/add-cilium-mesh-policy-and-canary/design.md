## Context

`replace-istio-with-cilium` removes Istio and leaves marketplace east–west as ordinary ClusterIP traffic plus a Cilium Gateway edge. The original Istio observe baseline listed strict mTLS, AuthorizationPolicy, retries, circuit breakers, and traffic splitting as explicit non-goals. This change is that list, plus Argo canary, implemented on Cilium rather than Istio CRDs.

Do not start implementation until Cilium Gateway is the live shop/pay origin.

## Goals / Non-Goals

**Goals:**

- Strict mTLS for marketplace service-to-service traffic (Cilium mutual auth / SPIFFE), fail closed for enrolled workloads.
- Authorization so only intended identities reach each gRPC/HTTP service.
- Retries and circuit breakers for at least the web → gRPC hop and/or Gateway → web path.
- Canary delivery for at least one marketplace workload via Argo Rollouts and Gateway API weight shifting on the existing Cilium HTTPRoutes.
- Documented rollback for policy (open up) and canary (abort to stable).

**Non-Goals:**

- Bringing Istio back.
- Canarying every service in the first slice (`web` is the default first target).
- Hubble L7 HTTP metrics dashboards (still optional).
- Kafka L7 policy or intercepting Strimzi TLS.
- Production-only features that staging has not proven.

## Decisions

### 1. Cilium policy, not Istio `PeerAuthentication` / `AuthorizationPolicy`

Use Cilium identity + `CiliumNetworkPolicy` (and mutual authentication mode required) for mTLS and authz. Map “who may call whom” from the existing service graph: `web` → users/products/orders/cart/payment; payment ↔ simulator as today; consumers from Kafka remain L4 to brokers in `kafka`.

**Alternatives considered:** Istio ambient waypoint policies (removed); Kubernetes NetworkPolicy only (no mTLS).

### 2. Retries and circuit breakers via Envoy on Cilium

Prefer Gateway API timeouts/retries where supported; otherwise `CiliumEnvoyConfig` / GAMMA HTTPRoutes for east–west. Do not add application-level retry storms in Go as a substitute.

**Alternatives considered:** client-side gRPC retries only (useful supplement, not the platform control).

### 3. Argo Rollouts + Gateway API plugin for canary

Install Argo Rollouts as a GitOps operator. Convert the first canary workload (likely `web`) from Deployment to Rollout. Shift traffic with HTTPRoute backend weights (stable vs canary Services) behind the existing Cilium Gateway.

**Rationale:** matches Gateway API already used for shop/pay; Argo CD is already the GitOps engine.

**Alternatives considered:** Argo CD blue/green with two Applications (heavier); Istio VirtualService weights (gone); Flagger (extra controller, overlapping).

### 4. Staging first, same Cloudflare path

Canaries and deny policies must be proven on staging (and Colima) without changing Cloudflare TLS. Failed canary must not require Cloudflare UI changes.

## Risks / Trade-offs

| Risk                                              | Mitigation                                                                                     |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| Fail-closed mTLS breaks CNPG, migrations, or jobs | Exclude data-plane exceptions (init/migrate, DB, Valkey) in the policy matrix before enforcing |
| Rollouts vs Tilt local Deployments                | Keep Colima on Deployment until Rollouts is documented for Tilt, or run Rollouts locally too   |
| Retry + checkout idempotency                      | Align retry budget with existing payment/order idempotency (do not blindly retry CreateOrder)  |
| Cilium mTLS vs Kafka                              | Keep kafka namespace free of required mTLS to marketplace pods that only use TLS to brokers    |

## Migration Plan

1. Land `replace-istio-with-cilium`.
2. Inventory allowed callers per Service; write observe-then-enforce policies.
3. Enable mutual auth in a non-enforcing mode if Cilium offers it; then required.
4. Add Rollouts operator; canary `web` with tiny weight; abort test.
5. Add retries/CB with conservative defaults; watch Hubble drops and app traces.
