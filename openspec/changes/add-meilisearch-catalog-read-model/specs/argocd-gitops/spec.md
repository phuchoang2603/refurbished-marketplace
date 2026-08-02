## ADDED Requirements

### Requirement: Meilisearch Application in app-of-apps

The shared Argo CD app-of-apps chart MUST define a child Application that deploys Meilisearch from the repository Helm chart/values under `infra/` for environments that enable platform search. Local and staging roots MUST be able to render that Application via app-of-apps values. Meilisearch SHALL deploy as a normal Helm workload (no ClickHouse/Elastic-style operator CRD prerequisite wave).

#### Scenario: Staging includes Meilisearch

- **WHEN** the staging root Application syncs with Meilisearch enabled in app-of-apps values
- **THEN** a child Application exists that deploys the Meilisearch Helm release into its configured namespace

#### Scenario: Local Argo can include Meilisearch

- **WHEN** the local root Application syncs with Meilisearch enabled in local app-of-apps values
- **THEN** a child Application exists for Meilisearch even though Tilt still owns the marketplace chart

#### Scenario: Master key is not committed in plaintext

- **WHEN** Meilisearch is deployed for staging or local GitOps
- **THEN** the Meilisearch master key SHALL be supplied via External Secrets or an equivalent secret reference rather than a committed plaintext value in Git
