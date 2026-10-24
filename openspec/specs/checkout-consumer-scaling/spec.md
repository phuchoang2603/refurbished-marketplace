# checkout-consumer-scaling Specification

## Purpose

Define production consumer-group and partition topology so shared checkout preparation can be processed concurrently across Orders and Inventory replicas and verified operationally.

## Requirements

### Requirement: Production preparation consumers use independent scalable groups

The production deployment SHALL configure three Orders replicas and three Inventory replicas. All Orders replicas SHALL share one stable Orders command consumer group; all Inventory reservation replicas SHALL share one stable Inventory command consumer group distinct from Orders. Both groups SHALL subscribe to `checkout.prepare.requested.v1`, which SHALL have six partitions in production and retain checkout identifiers as record keys. All other declared Kafka topics SHALL have one partition in both production and base configuration. Base configuration SHALL retain one replica per service and one preparation partition.

#### Scenario: One preparation command is published

- **WHEN** Checkout publishes one shared preparation record
- **THEN** both service groups SHALL independently receive that record, with one assigned member per partition within each stable group and possible redelivery during failures or rebalances

#### Scenario: Production Helm values are rendered

- **WHEN** the production overlays are rendered
- **THEN** Orders and Inventory SHALL each request three replicas, the shared preparation topic SHALL request six partitions, and all other Kafka topics SHALL request one partition

#### Scenario: Base Helm values are rendered

- **WHEN** the base charts are rendered without the production overlays
- **THEN** Orders and Inventory SHALL each request one replica and every Kafka topic SHALL request one partition

### Requirement: Parallel processing can be observed without weakening correctness

Operational verification SHALL expose consumer group membership, partition assignment, and processing evidence identifiable by checkout, service, and pod. A representative batch spanning all preparation partitions SHALL demonstrate that both groups consume the shared topic, all three healthy replicas per service receive assignments and process work, and Orders and Inventory processing can overlap. Verification SHALL NOT require an exact equal partition split or assume a latency improvement from replica count alone.

#### Scenario: Multiple checkout keys are processed

- **WHEN** a batch distributed across all six partitions is processed with three healthy members in each group
- **THEN** group descriptions and correlated logs or traces SHALL demonstrate assignments and completed work across the three replicas of each service, including overlapping processing across services

#### Scenario: One replica restarts

- **WHEN** an Orders or Inventory replica restarts while preparation is in progress
- **THEN** its group's partitions SHALL be reassigned and pending work SHALL resume without duplicate effective orders, stock holds, or gateway payments
