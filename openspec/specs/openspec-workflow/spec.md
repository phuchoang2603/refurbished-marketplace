# OpenSpec Workflow

## Purpose

OpenSpec is the repository's planning workflow for proposed work and active change tracking.

## Requirements

### Requirement: OpenSpec is the planning workflow

The repository MUST use OpenSpec change artifacts as the primary way to propose, design, and break down future work.

#### Scenario: New work is captured in OpenSpec

- **WHEN** a change is planned for the repository
- **THEN** the work SHALL be represented as an OpenSpec change instead of only in freeform planning notes

### Requirement: Change artifacts are authoritative for active work

Active change scope, design, and task breakdowns MUST live under `openspec/changes/` while a change is in progress.

#### Scenario: A change is being developed

- **WHEN** a developer or AI agent continues a change
- **THEN** the active artifacts SHALL be read from the matching OpenSpec change directory

### Requirement: Legacy planning notes are not the source of truth

`PLAN.md` and similar planning notes MUST NOT be treated as the authoritative source for new or in-progress work.

#### Scenario: A planning note disagrees with OpenSpec

- **WHEN** a planning note conflicts with OpenSpec change artifacts
- **THEN** the OpenSpec artifacts SHALL take precedence for the active change
