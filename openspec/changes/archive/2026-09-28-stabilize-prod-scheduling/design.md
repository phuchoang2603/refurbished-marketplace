## Context

See proposal.md. Observed on prod (`kubectl top`, 2026-09-28): `prod-server2` at 100% CPU while node requests were ~40%; brokers used 155–558m and Connect ~430m against 100m requests.

## Decisions

### 1. Dedicated `secret-store` chart at wave 1

A tiny chart renders only `SecretStore/doppler`; app-of-apps deploys it as `<env>-secret-store` at wave `1`, before MongoDB/Meilisearch (`2`), marketplace (`3`), and Kafka (`4`).

- Alternative: keep it in the marketplace chart. Rejected: marketplace is wave 3, after the MongoDB and Meilisearch ExternalSecrets that need it.
- Alternative: render it from the root app-of-apps. Rejected: the root's health would then depend on the Doppler token, hiding which layer failed.

### 2. Request sizing

Broker and Connect CPU requests go from 100m to 250m; memory requests are unchanged (observed working set is within them). Limits unchanged.

## Risks / Trade-offs

- [Higher broker requests (3×250m on prod) fail to fit] → Proxmox control-plane nodes have ~2.4 CPU of unrequested capacity each.

## Migration Plan

Merge to `main`; Argo CD creates `<env>-secret-store` and prunes the store from the three consumer Applications (ownership transfers via the tracking label). Dev needs SHA-tagged images for the new revision (dispatch Release Images). Rollback: revert the commit.
