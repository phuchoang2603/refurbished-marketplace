## ADDED Requirements

### Requirement: Progressive delivery operator in app-of-apps

The shared app-of-apps chart SHALL include an Argo Rollouts (or equivalent) operator Application ordered so it is Ready before marketplace canary Rollouts that depend on it.

#### Scenario: Rollouts syncs before canary workloads

- **WHEN** a full staging environment sync runs with canary enabled
- **THEN** the Rollouts controller Application is available before marketplace Rollout objects are relied upon
