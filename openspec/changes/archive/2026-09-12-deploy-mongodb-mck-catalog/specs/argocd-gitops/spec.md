## ADDED Requirements

### Requirement: MCK operator Application

The repository SHALL include an Argo CD child Application that deploys MongoDB Controllers for Kubernetes into the `operators` namespace at an operator sync wave (before marketplace and before the Mongo database Application).

#### Scenario: Operator syncs before the database CR

- **WHEN** a full environment sync runs
- **THEN** the MCK operator Application has a lower sync wave than the Application that applies the MongoDB Community custom resource

#### Scenario: Operator lands in operators

- **WHEN** the MCK operator Application syncs
- **THEN** the operator runs in the `operators` namespace

### Requirement: MongoDB Community Application in ecommerce

The repository SHALL include an Argo CD child Application that applies the MongoDB Community replica set (and related secrets/policy owned by that chart) into the `ecommerce` namespace. That Application SHALL NOT template an `ecommerce` Namespace object (the marketplace Application already destines that namespace).

#### Scenario: Database CR destines ecommerce

- **WHEN** the Mongo database Application syncs
- **THEN** the MongoDB Community custom resource is applied in `ecommerce`

#### Scenario: Prod may overlay size

- **WHEN** prod-root applies Mongo chart value overlays
- **THEN** production MAY raise replica-set members or resources without changing the Community operator path

### Requirement: GitOps docs include Mongo

GitOps documentation SHALL list the MCK operator and Mongo database Applications, their namespaces, and sync-wave order.

#### Scenario: Contributor finds Mongo in the deploy table

- **WHEN** a contributor reads the GitOps deploy guide
- **THEN** documentation names the Mongo operator and database Applications, `operators` vs `ecommerce`, and that shop traffic does not depend on Mongo yet
