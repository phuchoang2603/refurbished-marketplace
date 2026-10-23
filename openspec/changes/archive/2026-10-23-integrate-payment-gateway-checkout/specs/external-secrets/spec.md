## ADDED Requirements

### Requirement: Payment gateway credentials from Doppler

When the payment gateway integration is enabled, the chart SHALL render an ExternalSecret that syncs `PAYMENT_GATEWAY_API_KEY` and `PAYMENT_GATEWAY_WEBHOOK_SECRET` from Doppler into a Secret referenced only by Checkout. Environments without the integration SHALL render no such ExternalSecret.

#### Scenario: Prod gateway secret

- **WHEN** ExternalSecrets have synced in production
- **THEN** a `checkout-payment-gateway` Secret exists in `ecommerce` with both keys and only the Checkout Deployment mounts it

## MODIFIED Requirements

### Requirement: ExternalSecrets sync chart secrets

The repository SHALL render `ExternalSecret` resources from the `refurbished-marketplace` Helm chart for each enabled service with `db` (basic-auth username/password) and for each unique `auth.secretName` (for example `users-auth` / `JWT_SECRET`). Doppler remote password keys SHALL follow `{SECRET_NAME}_PASSWORD` derived from `db.secretName`. The Kubernetes Secret username SHALL be generated from `db.owner`, defaulting to `<service>_app`; no Doppler username key is required.

#### Scenario: DB secret available for CNPG

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `users-app` exists in `ecommerce` with `username` and `password` keys usable by CloudNativePG and service Helm templates

#### Scenario: JWT secret available for web and users

- **WHEN** ExternalSecrets have synced successfully
- **THEN** `users-auth` exists in `ecommerce` with `JWT_SECRET` key

#### Scenario: Debezium connector secrets

- **WHEN** ExternalSecrets have synced successfully
- **THEN** orders and inventory Postgres credentials and catalog Mongo credentials exist under the names referenced by Strimzi connectors; no obsolete products or payment Postgres credential is required
