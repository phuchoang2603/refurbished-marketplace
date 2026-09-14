## ADDED Requirements

### Requirement: Meilisearch Application in ecommerce

The repository SHALL include an Argo CD child Application that deploys Meilisearch into the `ecommerce` namespace. That Application SHALL NOT template an `ecommerce` Namespace object. Sync wave SHALL land after External Secrets Operator and with other data-plane stores, before marketplace workloads that query it at runtime.

#### Scenario: Search engine destines ecommerce

- **WHEN** the Meilisearch Application syncs
- **THEN** the Meilisearch workload is applied in `ecommerce`

#### Scenario: Search service can reference Meilisearch

- **WHEN** the marketplace Application syncs after Meilisearch is Healthy
- **THEN** the search workload can be configured with the in-cluster Meilisearch URL without a second search cluster

### Requirement: Search workload in marketplace chart

The marketplace Helm release SHALL deploy a `search` Service and Deployment in `ecommerce` with GHCR image tags matching other marketplace services.

#### Scenario: Search syncs with the chart

- **WHEN** the marketplace Application syncs
- **THEN** a search Deployment and Service exist in `ecommerce`

### Requirement: GitOps docs include Meilisearch

GitOps documentation SHALL list the Meilisearch Application, the search marketplace workload, namespaces, and sync-wave order relative to operators and marketplace.

#### Scenario: Contributor finds Meilisearch in the deploy table

- **WHEN** a contributor reads the GitOps deploy guide
- **THEN** documentation names the Meilisearch Application, the search service, `ecommerce`, and that Meilisearch is a catalog projection rather than listing SoT
