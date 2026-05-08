## Why

The current `PLAN.md` and loose `docs/` notes are drifting from the actual repository state, which makes them less useful as a source of truth. Moving to OpenSpec gives you a change-driven workflow where proposals, specs, and tasks stay closer to the codebase and are easier to keep current.

At the same time, you still need a stable place for big-picture diagrams and technical decisions. That should live in `docs/` as architecture notes and service-level decision records, not as mutable planning backlog.

## What Changes

- Introduce OpenSpec as the primary workflow for planning repo changes.
- Retire `PLAN.md` and move future change definition from planning notes into change-scoped artifacts under `openspec/changes/`.
- Establish `openspec/config.yaml` as the place for project context, conventions, and artifact rules.
- Capture repository-specific context such as the Go multi-module layout, REST-at-edge / gRPC-internal architecture, Kafka eventing, PostgreSQL/sqlc/goose conventions, and the new `devenv.nix`-managed development tooling.
- Keep `docs/` for durable architecture notes, Mermaid diagrams, ADR-style technical decisions, and service-level ownership notes.
- Reduce plan drift by keeping specs and change artifacts tied to the current repo structure and conventions.

## Capabilities

### New Capabilities

- `openspec-workflow`: adopt OpenSpec as the standard planning and change-tracking workflow for this repository.
- `project-context-config`: encode repo context and rules in `openspec/config.yaml` so artifact generation stays aligned with the current stack and conventions.
- `architecture-notes`: keep durable repo-level and service-level technical decisions in `docs/` without using them as a planning backlog.

### Modified Capabilities

- None.

## Impact

- `openspec/config.yaml` will be filled with project context and artifact rules.
- `PLAN.md` can be removed once its content has been migrated into OpenSpec and durable docs.
- Future planning will move away from `PLAN.md` and freeform docs toward OpenSpec change artifacts.
- `docs/` will shift from planning notes to stable architecture and decision documentation.
- The repo’s Go, Kafka, PostgreSQL, Kubernetes, and `devenv.nix` conventions will become explicit inputs to generated artifacts.
