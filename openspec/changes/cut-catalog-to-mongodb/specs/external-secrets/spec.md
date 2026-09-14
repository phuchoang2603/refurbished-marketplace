## ADDED Requirements

### Requirement: Products Mongo credentials from Doppler

The repository SHALL sync Doppler-backed credentials that products and the products Mongo outbox connector use to authenticate to the catalog replica set (SCRAM user password or equivalent). Plaintext Mongo URIs with passwords SHALL NOT be committed. Products SHALL NOT keep a Postgres `products-app` database secret after `products_db` is removed.

#### Scenario: Products can authenticate to Mongo

- **WHEN** ExternalSecrets have synced successfully after cutover
- **THEN** products and the Kafka Connect service account can read a Secret in `ecommerce` with the keys needed to connect to the catalog replica set

#### Scenario: Products Postgres secret is gone

- **WHEN** the marketplace chart no longer deploys `products_db`
- **THEN** GitOps SHALL NOT render a products CloudNativePG app username/password ExternalSecret
