## MODIFIED Requirements

### Requirement: Doppler environment configs

Dev SHALL use the marketplace dev Doppler config and prod prd. The operator SHALL use the matching platform kubeconfig to bootstrap a read-only marketplace token into ecommerce/doppler-token. CI SHALL NOT receive kubeconfig or application tokens.

#### Scenario: Bootstrap with kubectl

- **WHEN** an operator bootstraps an environment
- **THEN** its own ecommerce/doppler-token is bootstrapped privately from an environment-scoped read-only marketplace token without CI access

### Requirement: Doppler bootstrap secret manifests

The operator SHALL create the bootstrap Secret from an environment-scoped marketplace token. The repository SHALL document equivalent manual recovery without requiring committed token values.

#### Scenario: Developer creates the bootstrap secret

- **WHEN** an operator bootstraps an environment
- **THEN** its own ecommerce/doppler-token is bootstrapped privately from an environment-scoped read-only marketplace token without CI access

#### Scenario: Operator bootstraps prod cluster secret

- **WHEN** an operator bootstraps an environment
- **THEN** its own ecommerce/doppler-token is bootstrapped privately from an environment-scoped read-only marketplace token without CI access
