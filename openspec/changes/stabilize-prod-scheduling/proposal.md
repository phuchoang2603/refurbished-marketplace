## Why

After moving to per-environment Argo CD, prod's control-plane nodes are CPU-starved (Kafka brokers and Connect request 100m but use 400–560m), which stalls etcd and Argo CD, while the scheduler still sees ~60% free CPU so AWS burst workers never scale up. Separately, three charts each render `SecretStore/doppler`, so three Applications fight over one resource (`SharedResourceWarning`) and every one of them degrades when the store is unhealthy.

## What Changes

- A single `secret-store` child Application (new `infra/charts/secret-store`) owns `SecretStore/doppler` in `ecommerce`, synced before data stores; the `mongodb`, `meilisearch`, and `refurbished-marketplace` charts stop rendering it.
- Kafka broker and Kafka Connect CPU requests are raised to reflect observed usage so the scheduler spreads them and burst capacity becomes visible to the autoscaler.
- Marketplace and Kafka charts gain an opt-in AWS burst scheduling mode (`none` | `eligible` | `required`) that renders the talos-proxmox burst toleration and node affinity; default `none`.
- Prod marketplace service Deployments become `eligible` (prefer Proxmox, spill to AWS only when requests do not fit).
- Prod Kafka Connect becomes `required`, moving it to an AWS burst worker.

## Non-goals

- Changing the talos-proxmox autoscaler, ASG bounds, burst policy, or Strimzi operator placement.
- Moving any PVC-backed workload (CNPG, MongoDB, Meilisearch, Kafka brokers) to AWS.
- Right-sizing marketplace service requests (no usage data yet; revisit after they run).
- Changing image publishing triggers for infra-only commits.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `external-secrets`: The marketplace SecretStore is owned by exactly one Argo CD Application that syncs before its consumers.
- `argocd-gitops`: Marketplace workloads support opt-in AWS burst placement; prod marketplace services are burst-eligible and prod Kafka Connect runs on burst workers; Kafka requests reflect observed usage.

## Impact

- New `infra/charts/secret-store/`; `secret-store.tpl` removed from `infra/charts/{mongodb,meilisearch,refurbished-marketplace}`
- `infra/argocd/app-of-apps/values.yaml` gains the `secret-store` child
- `infra/charts/refurbished-marketplace/{values.yaml,values-prod.yaml,templates/services.tpl,templates/_helpers.tpl}`
- `infra/charts/kafka/{values.yaml,values-prod.yaml,templates/connect.tpl,templates/_helpers.tpl}`
- `docs/deployment/gitops.md`, `docs/development/secrets.md`
- Prod: one `m7i-flex.large` AWS worker stays up while Kafka Connect runs there.
