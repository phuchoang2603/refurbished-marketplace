## ADDED Requirements

### Requirement: Log lines follow the platform log contract

Marketplace JSON log lines SHALL be single-line JSON objects on stdout that use the top-level keys `time`, `level`, `msg`, `trace_id`, and `span_id` with slog's default meanings. The platform agent promotes those keys into the stored record's timestamp, severity, body, trace ID, and span ID. All other keys SHALL remain available as log attributes.

#### Scenario: Stored record has promoted fields

- **WHEN** a service logs with a context that has a valid span, and the platform agent's JSON parsing is deployed
- **THEN** the stored record has `TraceId` and `SpanId` from the line, `SeverityText` from `level`, `Body` from `msg`, and attributes such as `order_id` searchable in HyperDX

#### Scenario: Line without a span

- **WHEN** a service logs without a valid span
- **THEN** the stored record has an empty `TraceId` and still has its severity, body, and attributes

## MODIFIED Requirements

### Requirement: Logging documentation

The repository SHALL document structured logging field conventions, the JSON keys promoted by the platform agent, HyperDX log search examples filtered by service, trace ID, and checkout domain fields such as `order_id`, and trace-to-log navigation in `docs/deployment/observability.md`.

#### Scenario: Contributor finds logging guide

- **WHEN** a contributor opens observability documentation after this change
- **THEN** they can find the field table (including domain hot-path fields), a HyperDX search joining by trace ID or `order_id`, and trace-to-log steps
