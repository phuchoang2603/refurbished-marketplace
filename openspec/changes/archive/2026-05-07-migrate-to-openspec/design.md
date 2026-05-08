## Context

This repository currently uses `PLAN.md` and several `docs/` files to capture both active planning and long-lived architecture notes. That has started to drift from the actual codebase, especially as the project grew into a multi-module Go workspace with separate service boundaries, Kafka eventing, and a `devenv.nix`-managed tooling setup.

The goal is to move active planning into OpenSpec while keeping durable technical knowledge in `docs/`.

## Goals / Non-Goals

**Goals:**

- Make OpenSpec the source of truth for proposed work.
- Encode repository context and conventions in `openspec/config.yaml`.
- Preserve long-lived architecture notes, Mermaid diagrams, and technical decisions in `docs/`.
- Remove the planning backlog role from `PLAN.md` and plan-style docs.

**Non-Goals:**

- Rewriting the existing application architecture.
- Converting every doc into OpenSpec specs.
- Forcing all docs out of the repository.
- Implementing service changes beyond the documentation/workflow migration.

## Decisions

1. **OpenSpec owns active planning**
   - Proposed changes, requirements, design, and task breakdowns live under `openspec/changes/`.
   - Alternative considered: keep `PLAN.md` as the main planning doc.
   - Rejected because it already drifts from reality and lacks artifact boundaries.

2. **`docs/` becomes durable architecture documentation**
   - Keep one repo-level architecture note, plus per-service notes only where they add unique value.
   - Keep the Mermaid diagram in a single repo overview doc.
   - Keep ADR-style decisions and service ownership notes there.
   - Rename `docs/services/*-plan.md` files into stable service notes when they contain unique decisions, and delete them when they are just duplicated planning content.
   - Alternative considered: delete `docs/` entirely.
   - Rejected because the repo still needs a stable place for big-picture reasoning that is not a change proposal.

3. **`PLAN.md` is retired after migration**
   - Treat it as legacy planning content, not an ongoing source of truth.
   - Alternative considered: keep it as a high-level roadmap.
   - Rejected because it would continue to compete with OpenSpec and invite drift.

4. **Project context belongs in `openspec/config.yaml`**
   - Capture stack, domain, service boundaries, and repo rules there so new artifacts are generated against the real project shape.
   - Include the `devenv.nix`-managed tooling context so generated artifacts reflect the current developer workflow.
   - Alternative considered: duplicate context in every change proposal.
   - Rejected because it repeats the same assumptions and increases maintenance.

## Risks / Trade-offs

- Legacy docs may become stale if they are not clearly reclassified as architecture notes.
- `PLAN.md` removal could feel abrupt if the migration is not complete.
- Keeping docs and OpenSpec both in play requires a clear boundary between "decision record" and "proposed change".
- `openspec/config.yaml` can become too detailed if it starts duplicating actual specs or implementation notes.

## Migration Plan

1. Add the repo context and rules to `openspec/config.yaml`.
2. Create OpenSpec change artifacts for future work instead of extending `PLAN.md`.
3. Reframe `docs/` into one repo-level architecture note plus only the per-service notes that are actually needed.
4. Rename or delete `docs/services/*-plan.md` files depending on whether they contain unique decisions or duplicated planning content.
5. Remove or archive `PLAN.md` once its useful content has been migrated or superseded.

Rollback is simple: keep the existing docs if OpenSpec adoption stalls, but do not continue growing `PLAN.md` as the planning source of truth.

## Open Questions

- None.
