## ADDED Requirements

### Requirement: Marketplace chart deploys Meilisearch companion

The `refurbished-marketplace` Helm chart MUST be able to deploy a Meilisearch companion workload (Deployment, Service, and persistent volume) into the marketplace release namespace so local Tilt and the staging Argo marketplace Application deliver the catalog search read-model process with the rest of the marketplace. Meilisearch SHALL NOT be required to run as a per-pod sidecar on the products Deployment.

#### Scenario: Staging marketplace includes Meilisearch

- **WHEN** the staging marketplace Application syncs with Meilisearch enabled in chart values
- **THEN** a Meilisearch Deployment and Service exist in the marketplace namespace and expose the Meilisearch HTTP API for in-cluster clients

#### Scenario: Local Tilt marketplace includes Meilisearch

- **WHEN** Tilt deploys the marketplace chart with Meilisearch enabled
- **THEN** Meilisearch is available in-cluster to the products service without a separate app-of-apps Meilisearch Application

#### Scenario: Master key is not committed in plaintext

- **WHEN** Meilisearch is deployed for staging or local GitOps/Tilt
- **THEN** the Meilisearch master key SHALL be supplied via External Secrets or an equivalent secret reference rather than a committed plaintext value in Git

#### Scenario: Products does not embed Meilisearch as a sidecar

- **WHEN** the products Deployment is rendered
- **THEN** it SHALL reach Meilisearch over the in-cluster Service address and SHALL NOT require a Meilisearch container in the products pod
