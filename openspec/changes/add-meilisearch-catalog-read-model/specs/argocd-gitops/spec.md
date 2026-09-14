## ADDED Requirements

### Requirement: Meilisearch Application in ecommerce

The repository SHALL include an Argo CD child Application that deploys Meilisearch into the `ecommerce` namespace. That Application SHALL NOT template an `ecommerce` Namespace object. Sync wave SHALL land after External Secrets Operator and with other data-plane stores, before marketplace workloads that query it at runtime.

#### Scenario: Search destines ecommerce

- **WHEN** the Meilisearch Application syncs
- **THEN** the Meilisearch workload is applied in `ecommerce`

#### Scenario: Marketplace can reference search

- **WHEN** the marketplace Application syncs after Meilisearch is Healthy
- **THEN** products can be configured with the in-cluster Meilisearch URL without a second search cluster

### Requirement: GitOps docs include Meilisearch

GitOps documentation SHALL list the Meilisearch Application, namespace, and sync-wave order relative to operators and marketplace.

#### Scenario: Contributor finds Meilisearch in the deploy table

- **WHEN** a contributor reads the GitOps deploy guide
- **THEN** documentation names the Meilisearch Application, `ecommerce`, and that it is a catalog projection rather than listing SoT
