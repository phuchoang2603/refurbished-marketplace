## Purpose

Deliver the marketplace independently to Proxmox dev and AWS prod with matching published images, isolated secrets and verifiable GitOps convergence.

## ADDED Requirements

### Requirement: Trusted environment delivery

Trusted repository PRs targeting main SHALL deploy dev after chart validation and all required images publish. Main SHALL deploy prod. Fork PRs SHALL receive no deployment secrets. CI SHALL publish an environment deployment ref whose app-of-apps values pin all children to the immutable built source revision and image tag, without affecting the other environment. Each locally bootstrapped root SHALL watch its own deployment ref; CI SHALL NOT require Kubernetes credentials or network access.

#### Scenario: PR deployment

- **WHEN** all PR images publish successfully
- **THEN** dev reconciles that PR revision without changing prod

#### Scenario: Main deployment

- **WHEN** a main release publishes successfully
- **THEN** prod reconciles that main revision

### Requirement: Environment-specific platform consumption

Marketplace SHALL retain its own ambient-enrolled ecommerce namespace and consume existing platform operators, default storage and telemetry. Databases SHALL initially use one instance and ordinary services one replica. Dev SHALL publish shop-dev.phuchoang.sbs and prod shop.phuchoang.sbs through their own platform tunnels.

#### Scenario: Both environments render

- **WHEN** the shared charts render for dev and prod
- **THEN** both use the same application topology, their own default storage, and distinct public hostnames

### Requirement: Dev checkout remains unavailable

Dev SHALL omit payment gateway configuration and credentials until a separate gateway rollout. Marketplace readiness SHALL validate browsing, accounts, cart, search, persistence and event delivery without claiming payment acceptance.

#### Scenario: Dev checkout submission

- **WHEN** a user attempts checkout without gateway configuration
- **THEN** checkout is unavailable without creating an order or stock reservation
