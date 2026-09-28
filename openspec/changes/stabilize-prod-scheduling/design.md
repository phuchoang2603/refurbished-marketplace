## Context

See proposal.md. Observed on prod (`kubectl top`, 2026-09-28): `prod-server2` at 100% CPU while node requests were ~40%; brokers used 155–558m and Connect ~430m against 100m requests. The talos-proxmox burst contract: AWS workers carry label `burst.talos.dev/compute=aws` and taint `burst.talos.dev/stateless=true:NoSchedule`; the `burst-stateless-only` admission policy rejects PVC workloads that tolerate the taint; the ASG is 0–2 `m7i-flex.large` (1950m / 7274Mi allocatable).

## Decisions

### 1. Dedicated `secret-store` chart at wave 1

A tiny chart renders only `SecretStore/doppler`; app-of-apps deploys it as `<env>-secret-store` at wave `1`, before MongoDB/Meilisearch (`2`), marketplace (`3`), and Kafka (`4`).

- Alternative: keep it in the marketplace chart. Rejected: marketplace is wave 3, after the MongoDB and Meilisearch ExternalSecrets that need it.
- Alternative: render it from the root app-of-apps. Rejected: the root's health would then depend on the Doppler token, hiding which layer failed.

### 2. Burst mode as a small enum with a per-chart helper

Each chart gets a `burst.mode` default plus per-workload override (`services.<name>.burst`, `connect.burst`). A `_helpers.tpl` template renders `tolerations`/`affinity` for `eligible` and `required`. Taint/label keys are constants in the helper because they are the talos-proxmox contract, not tuning.

- Alternative: free-form `tolerations`/`affinity` passthrough values. Rejected: every overlay would restate the taint and label and could drift from the contract.

`eligible` uses a weight-100 preferred affinity for `burst.talos.dev/compute NotIn [aws]` so the scheduler keeps pods on Proxmox until requests no longer fit; the autoscaler then adds a worker for Pending pods. Migration Jobs and sidecar-only concerns are unchanged. Cart's Valkey sidecar has no volumes, so cart is eligible.

### 3. Connect `required` on prod via `template.pod`

Strimzi `KafkaConnect.spec.template.pod` carries tolerations and affinity. Connect is stateless (offsets/configs/status live in Kafka topics), so the admission policy allows it. Dev keeps `none` (single control plane; no need to pay for AWS).

### 4. Request sizing

Broker and Connect CPU requests go from 100m to 250m; memory requests are unchanged (observed working set is within them). Limits unchanged.

## Risks / Trade-offs

- [Connect on AWS adds WAN latency to CDC and pays a permanent AWS node] → CDC is asynchronous; one `m7i-flex.large` stays up while Connect runs. Revert with `connect.burst: none` in `values-prod.yaml`.
- [Cross-site MTU/fragmentation] → Cilium already pins the underlay MTU (validated in the burst-worker change); Connect uses TCP to brokers/DBs.
- [Autoscaler or ASG failure leaves Connect Pending] → CDC pauses but request paths are unaffected; check autoscaler logs per talos-proxmox burst operations doc.
- [Higher broker requests (3×250m on prod) fail to fit] → Proxmox control-plane nodes have ~2.4 CPU of unrequested capacity each.

## Migration Plan

Merge to `main`; Argo CD creates `<env>-secret-store`, prunes the store from the three consumer Applications (ownership transfers via the tracking label), and reschedules Connect on prod. Dev needs SHA-tagged images for the new revision (dispatch Release Images). Rollback: revert the commit.
