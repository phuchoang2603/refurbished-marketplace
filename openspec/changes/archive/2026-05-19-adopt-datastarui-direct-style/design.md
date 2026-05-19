## Context

The `services/web` module already uses the same underlying presentation stack as DatastarUI: `templ`, Tailwind CSS, and Datastar. It currently implements its own lightweight UI primitive layer in `internal/views/shared/ui.templ` and `ui_models.go`, while page templates in `internal/views/auth`, `cart`, `orders`, and `products` consume those shared wrappers.

The desired change is not to keep refining those local primitives. Instead, the web service will reset its UI foundation around vendored DatastarUI component and utility packages, while preserving the app's existing server-rendered request flow, shared shell, and Datastar-based fragment update model. Because this touches component package layout, Tailwind/theming assumptions, and all existing page templates, a design document is warranted before implementation.

## Goals / Non-Goals

**Goals:**

- Replace the current local primitive UI layer with vendored DatastarUI-based components and utilities.
- Adopt direct component usage style in page templates instead of preserving the current single shared wrapper package.
- Preserve the current HTML-first browser model, Datastar fragment responses, and existing route/handler boundaries.
- Keep app-specific shell, navigation, and page composition local to `services/web` while shifting generic components to the vendored foundation.
- End with the legacy `shared/ui.*` layer removed from the web service.

**Non-Goals:**

- Cloning the entire DatastarUI repository, including its demo pages, server entrypoint, or unrelated app structure.
- Rewriting web handlers, browser routes, or gRPC client composition outside what is required to update imports and templates.
- Introducing browser-facing JSON APIs or replacing Datastar-driven interactions with a different client-side model.
- Guaranteeing automatic compatibility with every upstream DatastarUI component; only selected vendored packages are in scope.

## Decisions

### Decision: Vendor selected DatastarUI packages into the web service instead of adding a direct module dependency

The implementation will copy only the reusable component and utility packages needed by `services/web` into a service-local package area, rather than importing `github.com/coreycole/datastarui` directly or cloning the full upstream repository.

Rationale:

- The upstream repo is a full application-shaped module, not a narrowly packaged UI SDK.
- Vendoring gives the web service a stable local copy without relying on an untagged upstream dependency.
- It keeps component ownership explicit inside this repo and allows small local adaptations when integration issues appear.

Alternatives considered:

- Add DatastarUI as a Go module dependency. Rejected because the upstream repo has no release or tag discipline yet, which makes dependency upgrades and breakage management harder.
- Clone the entire repo into `services/web`. Rejected because it would introduce duplicate app structure, pages, and server concerns that do not belong in this service.

### Decision: Adopt DatastarUI package shape directly in templates instead of recreating a compatibility wrapper layer

The page templates will import vendored component packages directly, following DatastarUI's component-per-package model, rather than recreating the current `shared.Button`, `shared.Card`, and similar APIs as a local compatibility facade.

Rationale:

- The user explicitly wants to stop maintaining a custom UI abstraction layer.
- Direct imports make the new component system the source of truth instead of hiding it behind another wrapper.
- It reduces the risk of half-migrating into a permanent compatibility layer that still requires local component design decisions.

Alternatives considered:

- Keep a shared wrapper package that adapts DatastarUI to existing page call sites. Rejected because it preserves the local abstraction the user wants to eliminate.

### Decision: Keep app shell and app-specific view helpers local while replacing generic primitives

The `AppShell`, navigation model, marketplace-specific view models, and route-specific page composition will remain local to `services/web`, while generic primitives and reusable interaction components move to the vendored foundation.

Rationale:

- Shell and navigation are application-specific, not generic design-system primitives.
- This limits the migration to the UI foundation without conflating it with app structure or business rendering logic.
- Existing handler and fragment targeting behavior already depends on stable page-level IDs and app-specific layout choices.

Alternatives considered:

- Replace shell and all view helpers with upstream layout patterns. Rejected because the current app shell is already aligned with the web service's routing and navigation needs.

### Decision: Migrate pages in-place, then delete the legacy shared UI layer after all call sites move

Implementation will introduce the vendored foundation first, port existing page templates to it, and only then remove `internal/views/shared/ui.templ` and related helper files.

Rationale:

- The existing shared primitives are load-bearing until all page templates are migrated.
- This sequencing avoids a broken intermediate state where the old layer is deleted before the new one is wired through every page.

Alternatives considered:

- Delete the old UI layer first and rebuild pages afterward. Rejected because it creates avoidable breakage and makes verification harder.

### Decision: Align Tailwind and theme foundations with the vendored components as part of the migration

The web service will update its Tailwind configuration and shared CSS foundation as needed to support the vendored DatastarUI components, including any token, utility, or plugin conventions required for correct rendering.

Rationale:

- The current Tailwind setup is minimal and does not yet expose a shadcn-style token layer.
- Many DatastarUI components assume shared visual tokens and richer component styling conventions than the current app defines.

Alternatives considered:

- Keep the current CSS foundation untouched and only copy template files. Rejected because interactive or styled components would likely render inconsistently or incompletely.

### Decision: Keep vendored package layout close to upstream DatastarUI structure, with only small local namespace adjustments when needed

The vendored component and utility packages will follow DatastarUI's project structure and naming as closely as practical inside `services/web`, while allowing small local namespace adjustments only where the web service needs clearer ownership or import paths.

Rationale:

- The user prefers to adopt the DatastarUI project structure as much as possible.
- Staying close to upstream makes future sync work and source comparison easier.
- Small namespace adjustments still leave room to keep the vendored code understandable inside this repo.

Alternatives considered:

- Fully redesign the package layout around existing `services/web` naming. Rejected because it would turn vendoring into a local rewrite and make upstream comparison harder.

### Decision: Replace the current marketplace-specific visual styling rather than preserving it during the first migration

The first migration will discard the current custom marketplace UI styling in favor of the vendored DatastarUI visual foundation, keeping only app-specific structure such as routes, shell responsibilities, and navigation semantics.

Rationale:

- The user explicitly wants to discard the old UI rather than preserve or blend it.
- A full visual reset reduces hybrid styling states where old and new primitives coexist awkwardly.
- This keeps the migration aligned with the goal of adopting DatastarUI as the source of truth for presentation.

Alternatives considered:

- Preserve parts of the existing shell styling and gradually blend them with DatastarUI defaults. Rejected because it would prolong the old design system and weaken the reset.

### Decision: Discover the minimum required DatastarUI package set by tracing current page usage, while preferring upstream structure over ad hoc local abstractions

The implementation will start from the primitives and interactions currently used by existing pages, identify the minimum DatastarUI component and utility packages needed to replace them, and vendor those packages in an upstream-like layout.

Rationale:

- The exact package set is not known yet from the current codebase alone.
- Starting from actual page usage keeps the initial vendored footprint minimal.
- Combining that discovery step with an upstream-like structure avoids premature reorganization.

Alternatives considered:

- Vendor a broad swath of DatastarUI packages up front. Rejected because it increases migration surface area before the current app proves it needs those components.

## Risks / Trade-offs

- [Vendored upstream code will drift from DatastarUI over time] -> Mitigate by vendoring only selected packages, recording the upstream source commit during implementation, and treating future upgrades as explicit sync work.
- [Direct component package imports will increase page-level churn] -> Mitigate by migrating route groups systematically and keeping app-specific helpers only where they materially reduce duplication.
- [Tailwind and theme changes can create broad visual regressions] -> Mitigate by verifying all existing browser pages after the CSS foundation is updated and adjusting the shell only where needed for consistency.
- [Some DatastarUI components may assume supporting helpers not obvious from the first copied files] -> Mitigate by tracing full component dependencies before vendoring and importing the needed utility packages together rather than one file at a time.

## Migration Plan

1. Identify the DatastarUI components and utility packages required to support the existing marketplace pages.
2. Vendor those packages into `services/web` under a service-local package layout that preserves direct component usage.
3. Update Tailwind configuration and shared CSS/theme setup to support the vendored component foundation.
4. Migrate server-rendered page templates and shell references from the legacy shared primitives to the vendored packages.
5. Remove the old `internal/views/shared/ui.*` layer and any unused helper code after all page call sites compile and render correctly.
6. Verify full-page rendering plus Datastar fragment behaviors across auth, catalog, cart, and orders flows.

Rollback strategy:

- Revert the vendored package introduction and page migration together if the new component foundation causes broad regressions.
- Because this change stays inside `services/web` presentation code, rollback does not require data migration or service-contract rollback.

## Open Questions

- None.
