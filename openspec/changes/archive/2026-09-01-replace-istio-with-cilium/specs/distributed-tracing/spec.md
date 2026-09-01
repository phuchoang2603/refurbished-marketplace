## MODIFIED Requirements

### Requirement: Tracing documentation

The repository SHALL document the end-to-end tracing architecture, TraceId joining rules for application and async hops, outbox/Debezium configuration, operation-centric naming expectations, DB/Redis child spans, and Grafana verification steps that do not depend on Gateway or Hubble proxy spans.

#### Scenario: Contributor finds the tracing guide

- **WHEN** a contributor opens observability documentation after this change
- **THEN** they can follow steps to locate a checkout TraceId and interpret app, DB/Redis, and async spans without expecting mesh or Gateway proxy spans in the default view
