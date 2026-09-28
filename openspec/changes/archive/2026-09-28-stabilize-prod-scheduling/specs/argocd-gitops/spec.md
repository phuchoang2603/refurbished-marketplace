## ADDED Requirements

### Requirement: Kafka requests reflect observed usage

Kafka broker and Kafka Connect CPU requests SHALL be sized to their observed steady-state usage rather than a nominal minimum, so the scheduler sees their real load.

#### Scenario: Broker and Connect requests

- **WHEN** the kafka chart renders with default values
- **THEN** broker and Connect containers request at least 250m CPU
