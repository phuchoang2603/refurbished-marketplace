## Purpose

Define GitOps-managed canary deployments for marketplace workloads using Argo Rollouts and Gateway API traffic splitting on the Cilium edge.

## ADDED Requirements

### Requirement: Argo Rollouts is GitOps-managed

The repository SHALL deploy Argo Rollouts (or an equivalent Argo progressive-delivery controller) through the shared app-of-apps chart for local and staging.

#### Scenario: Root sync includes Rollouts

- **WHEN** the local or staging root Application syncs from Git after this change
- **THEN** Argo CD manages a child Application for the Rollouts controller

### Requirement: At least one marketplace canary

At least one marketplace workload (default: `web`) SHALL be delivered as an Argo Rollout that can shift traffic between stable and canary using Gateway API HTTPRoute backend weights on the existing Cilium Gateway.

#### Scenario: Canary weight is shifted

- **WHEN** a new `web` image is rolled out with a canary step
- **THEN** a fraction of shop hostname traffic reaches the canary Service without changing Cloudflare Public Hostnames

#### Scenario: Canary abort restores stable

- **WHEN** a canary is aborted
- **THEN** shop traffic returns to the stable Service without application code changes

### Requirement: Canary is staging-proven

Production canary enablement SHALL remain opt-in until staging has completed a successful canary and abort test.

#### Scenario: Production not implicit

- **WHEN** production manifests are rendered before canary enablement is chosen
- **THEN** production marketplace workloads are not converted to Rollouts by accident
