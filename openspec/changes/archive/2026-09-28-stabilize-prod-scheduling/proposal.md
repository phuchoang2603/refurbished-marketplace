## Why

After moving to per-environment Argo CD, prod's control-plane nodes are CPU-starved (Kafka brokers and Connect request 100m but use 400–560m), which stalls etcd and Argo CD while the scheduler still sees ~60% free CPU. Separately, three charts each render `SecretStore/doppler`, so three Applications fight over one resource (`SharedResourceWarning`) and every one of them degrades when the store is unhealthy.

## What Changes

- A single `secret-store` child Application (new `infra/charts/secret-store`) owns `SecretStore/doppler` in `ecommerce`, synced before data stores; the `mongodb`, `meilisearch`, and `refurbished-marketplace` charts stop rendering it.
- Kafka broker and Kafka Connect CPU requests are raised to reflect observed usage so the scheduler spreads them across nodes.

## Non-goals

- Scheduling marketplace workloads on talos-proxmox AWS burst workers.
- Right-sizing marketplace service requests (no usage data yet; revisit after they run).
- Changing image publishing triggers for infra-only commits.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `external-secrets`: The marketplace SecretStore is owned by exactly one Argo CD Application that syncs before its consumers.
- `argocd-gitops`: Kafka requests reflect observed usage.

## Impact

- New `infra/charts/secret-store/`; `secret-store.tpl` removed from `infra/charts/{mongodb,meilisearch,refurbished-marketplace}`
- `infra/argocd/app-of-apps/values.yaml` gains the `secret-store` child
- `infra/charts/kafka/values.yaml`
- `docs/deployment/gitops.md`, `docs/development/secrets.md`
