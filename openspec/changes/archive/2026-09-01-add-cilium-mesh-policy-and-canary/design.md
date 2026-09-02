## Context

See proposal.md for motivation. `replace-istio-with-cilium` left marketplace east–west as ClusterIP plus a Cilium Gateway edge. Cilium Helm lives in talos-proxmox; this repo must not helm-upgrade the CNI. Hubble is off on Talos. Argo CD is the only applier.

## Goals / Non-Goals

**Goals:**

- Fail-closed Cilium mutual auth + identity allow-lists for enrolled marketplace hops, with a documented exception matrix.
- HTTPRoute timeouts on Gateway → web / pay; no dataplane retries until checkout idempotency (#35).
- Rollback is Git + Argo sync (policy off).

**Non-Goals:**

- Canary, traffic splitting, Argo Rollouts, Flagger.
- Installing SPIRE from this marketplace repo.
- `CiliumEnvoyConfig` for east–west or explicit circuit-breaker CRDs.
- HTTPRoute or gRPC retries on checkout/payment paths.

## Decisions

### 1. CiliumNetworkPolicy + mutual auth, not Istio CRDs

Use Cilium identities and `CiliumNetworkPolicy` (ingress allow-lists; `authentication.mode: required` on enrolled hops; `enableDefaultDeny.ingress: true` when `meshPolicy.enforce` is true). Map callers: `web` → users/products/orders/cart/payment; simulator → web callback; Kafka stays L4/TLS to `kafka`.

SPIRE is enabled in talos-proxmox Cilium Helm before `meshPolicy.mutualAuth: true`.

### 2. HTTPRoute timeouts only; outlier detection from Cilium Gateway

Set `timeouts.request` / `timeouts.backendRequest` on shop and pay HTTPRoutes. Do not add HTTPRoute `retry` filters or gRPC retry interceptors — `CreateOrder` and hosted-payment callbacks are not idempotent at this layer ([#35](https://github.com/phuchoang2603/refurbished-marketplace/issues/35) is the prerequisite for safe checkout retries).

Cilium Gateway applies Envoy outlier detection on backend clusters by default; this repo does not add `CiliumEnvoyConfig`.

### 3. No canary in this change

Shop traffic stays on the single `web` Service and HTTPRoute backend.

### 4. talos-dev verification before archive

Prove deny (probe pod), allowed path (shop browse + checkout), and SPIRE Ready on talos-dev.

## Risks / Trade-offs

| Risk                                              | Mitigation                                              |
| ------------------------------------------------- | ------------------------------------------------------- |
| Fail-closed mTLS breaks CNPG, migrations, or jobs | Exception matrix; CNPs select app pods only             |
| Required mTLS before SPIRE                        | Enable SPIRE in talos-proxmox first                     |
| No dataplane retries until #35                    | Timeouts + implicit outlier detection only              |
| `enforce: false` without default deny             | `enableDefaultDeny.ingress: false` only in observe mode |

## Migration Plan

1. Confirm Cilium Gateway and SPIRE on talos-dev.
2. Sync marketplace chart with `meshPolicy.enabled: true`, `enforce: true`, `mutualAuth: true`.
3. Run probe deny + checkout on talos-dev.
4. Watch agent logs and shop traces after sync.
