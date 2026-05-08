## 1. OpenSpec setup

- [x] 1.1 Fill `openspec/config.yaml` with repo context, conventions, and `devenv.nix` tooling details
- [x] 1.2 Verify OpenSpec change artifacts for this migration are complete and internally consistent

## 2. Docs migration

- [x] 2.1 Create a single repo-level architecture overview doc in `docs/` with the Mermaid diagram
- [x] 2.2 Rename `docs/services/*-plan.md` files that contain unique decisions into stable service notes
- [x] 2.3 Delete `docs/services/*-plan.md` files that only duplicate planning content already captured elsewhere

## 3. Legacy plan removal

- [x] 3.1 Migrate any useful content from `PLAN.md` into OpenSpec or durable docs
- [x] 3.2 Remove or archive `PLAN.md` after the migration is complete

## 4. Validation

- [x] 4.1 Confirm the docs boundary is clear: OpenSpec for change proposals, `docs/` for stable architecture notes
- [x] 4.2 Review the repository for stale references to `PLAN.md` or planning-style docs
