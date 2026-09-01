## Context

See proposal.md for motivation. `replace-istio-with-cilium` left marketplace east–west as ClusterIP plus a Cilium Gateway edge. Cilium Helm lives in talos-proxmox; this repo must not helm-upgrade the CNI. Hubble is off on Talos. Argo CD is the only applier. The original Istio observe baseline deferred mTLS, authz, retries, and circuit breakers — this change implements that list on Cilium. Progressive delivery / canary is explicitly out of scope.

## Goals / Non-Goals

**Goals:**

- Fail-closed Cilium mutual auth + identity allow-lists for enrolled marketplace hops, with a documented exception matrix.
- Conservative retries/CB on Gateway → web and/or web → gRPC without retrying CreateOrder / payment callbacks unsafely.
- Rollback is Git + Argo sync (policy off).

**Non-Goals:**

- Canary, traffic splitting, weighted HTTPRoutes, Argo Rollouts, Flagger.
- Installing SPIRE or flipping Cilium Helm from this marketplace repo (operator enables it in talos-proxmox).
- Re-enabling Hubble as a requirement.

## Decisions

### 1. CiliumNetworkPolicy + mutual auth, not Istio CRDs

Use Cilium identities and `CiliumNetworkPolicy` (ingress/egress allow-lists; `authentication.mode: required` on enrolled hops). Map callers from the existing graph: `web` → users/products/orders/cart/payment; payment ↔ simulator; consumers to Kafka stay L4/TLS to `kafka`.

**Cluster dependency:** Cilium mutual authentication needs SPIRE via Cilium Helm (`authentication.mutual.spire.enabled`) in talos-proxmox `apps/values/cilium.yaml`. The operator will enable that there. This change assumes it is present before enforcing `authentication.mode: required`.

**Alternatives considered:** Istio ambient waypoint policies (removed); Kubernetes NetworkPolicy only (no mTLS).

### 2. Retries and circuit breakers via Gateway API / Cilium Envoy

Prefer HTTPRoute timeouts/retries on the shop Gateway path. For circuit breaking, use `CiliumEnvoyConfig` / `CiliumClusterwideEnvoyConfig` as Cilium documents for Envoy circuit breakers. Do not add application-level retry storms in Go as the platform control.

**Alternatives considered:** client-side gRPC retries only (supplement, not the control plane).

### 3. No canary in this change

Shop traffic stays 100% on the existing `web` Service. Do not add a second Deployment/Service or HTTPRoute weights.

**Rationale:** operator dropped progressive delivery from this slice.

**Alternatives considered:** GitOps HTTPRoute weights; Argo Rollouts (both rejected).

### 4. Staging first, same Cloudflare path

Prove deny policies and required mTLS on talos-dev without changing Cloudflare TLS or Public Hostnames.

## Risks / Trade-offs

| Risk                                              | Mitigation                                                                                    |
| ------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Fail-closed mTLS breaks CNPG, migrations, or jobs | Exception matrix before enforce (init/migrate, DB, Valkey, kafka)                             |
| Required mTLS before SPIRE is live on Talos       | Land talos-proxmox Cilium Helm first; do not flip `authentication.mode: required` until it is |
| Retry + checkout idempotency                      | Do not blindly retry CreateOrder / hosted-payment callbacks                                   |
| Hubble off                                        | Verify with `cilium` policy verdicts / agent logs and app traces, not a Hubble UI requirement |

## Migration Plan

1. Confirm Cilium Gateway is live; confirm talos-proxmox has mutual-auth/SPIRE.
2. Inventory allowed callers; observe-then-enforce CNPs.
3. Enable `authentication.mode: required` on enrolled hops.
4. Add retries/CB with conservative defaults; watch drops and traces.
