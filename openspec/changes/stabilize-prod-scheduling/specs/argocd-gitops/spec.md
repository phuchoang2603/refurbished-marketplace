## ADDED Requirements

### Requirement: Opt-in AWS burst placement

The `refurbished-marketplace` and `kafka` charts SHALL support a burst mode per workload with values `none`, `eligible`, and `required`, defaulting to `none`. `eligible` SHALL render the `burst.talos.dev/stateless=true:NoSchedule` toleration and a preferred node affinity away from `burst.talos.dev/compute=aws`. `required` SHALL render the toleration and a required node affinity for `burst.talos.dev/compute=aws`. Burst modes SHALL only be applied to workloads without persistent volumes; migration Jobs and PVC-backed workloads SHALL NOT receive the toleration.

#### Scenario: Default stays on Proxmox

- **WHEN** a chart renders with default values
- **THEN** no workload carries the burst toleration or burst node affinity

#### Scenario: Eligible workload prefers Proxmox

- **WHEN** a service renders with burst mode `eligible`
- **THEN** its pod template tolerates the burst taint and prefers nodes without `burst.talos.dev/compute=aws`

#### Scenario: Required workload runs on AWS

- **WHEN** Kafka Connect renders with burst mode `required`
- **THEN** its pod template tolerates the burst taint and requires `burst.talos.dev/compute=aws`

### Requirement: Production burst placement

Production SHALL set marketplace service Deployments to burst mode `eligible` and Kafka Connect to `required`. Dev SHALL keep chart defaults.

#### Scenario: Prod Connect lands on a burst worker

- **WHEN** `prod-kafka` syncs
- **THEN** the Kafka Connect pod is scheduled on a node labeled `burst.talos.dev/compute=aws`, triggering an AWS scale-up if none exists

#### Scenario: Prod services spill only under pressure

- **WHEN** `prod-refurbished-marketplace` syncs and Proxmox nodes have room for the service requests
- **THEN** marketplace service pods schedule on Proxmox nodes

### Requirement: Kafka requests reflect observed usage

Kafka broker and Kafka Connect CPU requests SHALL be sized to their observed steady-state usage rather than a nominal minimum, so scheduling and autoscaling see their real load.

#### Scenario: Broker and Connect requests

- **WHEN** the kafka chart renders with default values
- **THEN** broker and Connect containers request at least 250m CPU
