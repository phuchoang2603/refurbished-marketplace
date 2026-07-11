# ArgoCD GitOps

## Purpose

Define staging and production GitOps delivery via ArgoCD app-of-apps, environment Helm overlays, and coordinated GHCR image tags without requiring Tilt on remote clusters.

## Requirements

### Requirement: App-of-apps per environment

The repository SHALL provide ArgoCD Application manifests under `infra/argocd/<environment>/` for `staging` and `production` only. Each environment SHALL include a root Application that syncs child Applications for operators, the `refurbished-marketplace` Helm chart, and the `kafka` Helm chart.

#### Scenario: Staging root application

- **WHEN** the staging cluster root Application syncs from Git
- **THEN** child Applications exist for operators, `refurbished-marketplace`, and `kafka` in the `ecommerce` and `operators` namespaces as defined

#### Scenario: No dev ArgoCD overlay

- **WHEN** a developer uses Tilt for local development
- **THEN** no `infra/argocd/values/dev/` or dev Application set is required; chart default `values.yaml` remains the dev source

### Requirement: Environment-specific Helm values

The repository SHALL provide Helm value overlays at `infra/argocd/values/staging/` and `infra/argocd/values/production/` for the marketplace and kafka charts. Staging overlays SHALL set `global.imageTag` to `main`. Production overlays SHALL set `global.imageTag` to a commit SHA for coordinated releases.

#### Scenario: Staging pulls rolling main tag

- **WHEN** the staging marketplace Application syncs
- **THEN** Helm values set `global.imageRegistry` to the project GHCR path and `global.imageTag` to `main`

#### Scenario: Production pins commit SHA

- **WHEN** the production marketplace Application syncs after a promotion
- **THEN** Helm values set `global.imageTag` to the promoted commit SHA shared by all service images

### Requirement: Chart image registry and tag resolution

The `refurbished-marketplace` and `kafka` Helm charts SHALL support `global.imageRegistry` and `global.imageTag`. When `global.imageRegistry` is empty, templates SHALL render service image fields unchanged for local Tilt. When set, templates SHALL render images as `{registry}/{shortName}:{tag}`.

#### Scenario: Tilt local image names unchanged

- **WHEN** Helm renders with default chart values and empty `global.imageRegistry`
- **THEN** service deployments reference short names such as `refurbished-marketplace/web`

#### Scenario: Remote cluster GHCR reference

- **WHEN** Helm renders with `global.imageRegistry` set and `global.imageTag` set to `main`
- **THEN** a service with `image: web` deploys as `ghcr.io/<repository>/web:main`

### Requirement: Payment gateway simulator in marketplace chart

The repository SHALL deploy `payment-gateway-simulator` from the `refurbished-marketplace` Helm chart. Staging and production value overlays SHALL configure `HOSTED_PAYMENT_BASE_URL` to reach the in-cluster simulator service.

#### Scenario: Simulator enabled in staging

- **WHEN** the staging marketplace chart syncs
- **THEN** a `payment-gateway-simulator` Deployment and Service exist in `ecommerce`

#### Scenario: Web uses in-cluster simulator URL remotely

- **WHEN** staging or production marketplace values are applied
- **THEN** the web service `HOSTED_PAYMENT_BASE_URL` targets the in-cluster simulator, not `localhost`

### Requirement: Loose sync ordering

Child ArgoCD Applications SHALL use sync waves so operators sync before the marketplace chart and the marketplace chart syncs before the kafka chart.

#### Scenario: Operator wave before apps

- **WHEN** a full environment sync runs
- **THEN** operator Applications have a lower sync wave than marketplace and kafka Applications

### Requirement: GitOps documentation

The repository SHALL document the ArgoCD layout, staging vs production value locations, image tag promotion for production, and prerequisites that remain outside Git (Argo bootstrap, Doppler token, ClusterSecretStore).

#### Scenario: Contributor finds deploy guide

- **WHEN** a contributor prepares a staging or production deploy
- **THEN** development documentation explains app-of-apps paths, value overlays, and SHA promotion without requiring Tilt
