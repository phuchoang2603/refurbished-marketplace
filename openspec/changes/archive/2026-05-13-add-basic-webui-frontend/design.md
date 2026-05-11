## Context

`services/web` already owns the public browser edge and renders `templ` pages/fragments with Datastar-compatible interactions. The current templates prove the architecture but are sparse: styling is mostly inline in `AppShell`, catalog/cart/order pages expose raw IDs, and auth responses are functional rather than product-like.

The desired direction is still server-rendered HTML in `services/web`, not a separate browser app under `clients/`. Internal service communication remains gRPC, and browser requests remain form/HTML-first.

## Goals / Non-Goals

**Goals:**

- Provide a basic cohesive marketplace UI for catalog, product detail, cart, auth, and order pages.
- Move shared visual styling into static assets served by `services/web`.
- Copy and adapt DatastarUI primitives broadly for compatible `templ` UI components without introducing a client-side application layer.
- Keep Datastar behavior embedded in templates and use `datastar-go` for server-side SSE/patch helpers where interactions need richer updates.
- Persist browser auth tokens in HTTP cookies rather than displaying raw tokens as the primary browser UX.
- Compose cart rows with product details through the existing products gRPC client.
- Improve user-facing labels and structure while keeping service-boundary changes limited to auth cookie expiry support.
- Add tests that verify rendered pages, static assets, and fragment targets.

**Non-Goals:**

- Creating a separate SPA or frontend package under `clients/`.
- Introducing Node, npm, bundlers, or client-side state management.
- Redesigning the full brand/visual identity.
- Changing persistence schemas or Kafka event flows.
- Making cart preview pricing authoritative for checkout or order totals.

## Decisions

- Keep the frontend inside `services/web`.
  - Rationale: the web service already owns browser routing, auth boundary, gRPC composition, and server-rendered responses.
  - Alternatives considered: create `clients/frontend`; this would duplicate the browser boundary and pull the project toward SPA/client API contracts that the current architecture intentionally avoids.

- Expose a single HTTP port from the web container.
  - Rationale: the same Go HTTP server serves HTML pages, static CSS/assets, Datastar interaction endpoints, and non-browser routes such as health checks. The browser talks only to `services/web`; `services/web` then composes internal services over gRPC from the same process.
  - Alternatives considered: run a separate frontend server/container or expose a second port for assets; this is unnecessary without a SPA/build server and would complicate Kubernetes routing.

- Serve plain static assets from `services/web/static/`.
  - Rationale: a single CSS file is enough for a basic UI and keeps build/deploy complexity low.
  - Alternatives considered: keep all CSS inline in `shell.templ`; this is simple but will grow hard to maintain. Add a CSS framework or bundler; that is premature for a basic UI pass.

- Copy DatastarUI primitives into the app instead of treating it as an installed component library.
  - Rationale: DatastarUI is inspired by shadcn/ui: it is a collection of reusable components to copy and adapt, not a dependency to hide behind. Copying primitives into `services/web/internal/views/` makes the UI code local, inspectable, and easy to modify while still using DatastarUI patterns broadly for buttons, forms, cards, empty states, navigation primitives, and interactive fragments.
  - Alternatives considered: install DatastarUI as a dependency; this would create unnecessary coupling if the intended workflow is copy-and-adapt. Plain CSS only; this misses the desired primitive reuse. TailwindCSS; defer unless the project intentionally adopts a utility-first frontend workflow.

- Use `datastar-go` selectively for Datastar server responses.
  - Rationale: full page requests should remain normal HTML responses, while interactive actions can use `datastar-go` to read signals and emit SSE patches, redirects, signal updates, or multi-target DOM updates without adding browser JSON endpoints.
  - Alternatives considered: keep custom Datastar headers and raw fragments only; simple for one-target patches but less expressive for richer interactions. Use custom SSE formatting; unnecessary if the SDK handles Datastar response semantics.

- Package static assets with the existing web container.
  - Rationale: `infra/docker/web.Dockerfile` already copies the full `services/web` directory in the builder stage. The runtime image currently only copies the compiled binary, so static files must be explicitly copied into the final image from `/src/services/web/static` to `/static`. The web service should use `/static` as the fixed runtime asset directory.
  - Alternatives considered: embed assets with `go:embed`; this keeps the runtime image simple but makes asset updates part of the binary and was not chosen for this pass. Serve assets from a separate frontend container; this is unnecessary for server-rendered web UI.

- Keep templates componentized by page and reusable sections.
  - Rationale: Datastar fragments need stable DOM targets, and `templ` components already provide a clean page/section split.
  - Alternatives considered: collapse markup into handlers; this would make UI changes harder and mix presentation with edge orchestration.

- Improve display models at the web edge, with only minimal service contract changes where browser session behavior requires it.
  - Rationale: the web service can shape gRPC data for presentation while preserving service boundaries. The one exception is refresh cookie lifetime: the web edge needs the users service refresh-token expiry to avoid inventing a separate cookie duration policy.
  - Alternatives considered: keep refresh cookies as session cookies; simpler but diverges from the users service token policy. Add unrelated domain fields or protobuf changes; not needed for a basic first UI.

- Compose cart rows with product details in `services/web`.
  - Rationale: product IDs alone are not usable for a basic marketplace cart. The web edge already composes data over gRPC and can enrich cart rows with product names and current catalog prices.
  - Alternatives considered: keep product-ID-only rows to minimize calls; this preserves simplicity but makes the UI feel incomplete. Add cart service product snapshots; that would change service contracts and persistence semantics.

- Persist auth tokens in HTTP cookies for browser flows.
  - Rationale: server-rendered protected pages should not require users to manually copy bearer tokens. Cookies fit normal browser navigation and form submissions. Access and refresh cookie lifetimes should use the expiry values returned by the users service directly so the web edge does not invent separate session duration policy. The users token response therefore includes refresh-token expiry in addition to access-token expiry.
  - Alternatives considered: keep displaying raw tokens; useful for debugging but poor UX and easy to leak. Introduce a separate session store; heavier than needed while users service already issues tokens. Configure independent cookie TTLs in web; this risks drifting from users service token policy.

- Keep interactions form-first with Datastar enhancements.
  - Rationale: forms preserve simple browser behavior, while Datastar upgrades selected flows with fragments.
  - Alternatives considered: add browser JSON APIs; this conflicts with the current web contract unless needed by a non-browser consumer.

## Risks / Trade-offs

- [Risk] The UI may outgrow hand-written CSS quickly.
  - Mitigation: keep the first stylesheet small and organized around layout, cards, forms, tables, and utility states.

- [Risk] Cart composition adds one products gRPC call per cart item.
  - Mitigation: keep the first implementation simple because cart sizes are expected to be small; revisit batching only if needed.

- [Risk] Cart preview prices can differ from final order totals if product prices change.
  - Mitigation: label cart totals as estimates/current catalog prices and keep order creation as the source of truth for final totals.

- [Risk] Datastar fragment targets can drift from template IDs.
  - Mitigation: add tests for expected target IDs and interaction attributes.

- [Risk] Cookie-backed auth introduces CSRF and environment-specific cookie security concerns.
  - Mitigation: use `HttpOnly`, `SameSite=Lax`, clear cookies on logout, and only use `Secure` where HTTPS is available.

- [Risk] Copied DatastarUI primitives can drift from upstream examples.
  - Mitigation: treat copied primitives as local application code and adapt intentionally; document any major deviations in component comments only when needed.

- [Risk] Static assets can be missing in the distroless runtime image if the Dockerfile is not updated.
  - Mitigation: update `infra/docker/web.Dockerfile` to copy `services/web/static` into `/static` in the final image.

- [Risk] `datastar-go` SSE responses may be overkill for simple one-target interactions.
  - Mitigation: keep normal HTML responses for full pages and use SDK-backed SSE only for interactive endpoints that benefit from patches, redirects, signal updates, or multiple DOM targets.

## Migration Plan

- Add static asset serving before referencing assets from the shell, using `/static` in the web container.
- Copy DatastarUI-inspired primitives into `services/web/internal/views/` and use them broadly for compatible server-rendered UI before filling gaps with custom local components.
- Add `datastar-go` for interactive response helpers and use it selectively for Datastar action endpoints.
- Add cookie read/write support before changing protected browser auth behavior.
- Update shell and page templates incrementally while preserving route behavior.
- Regenerate `templ` output after template changes.
- Run `go test ./...` in `services/web`.

Rollback is straightforward: revert template/static asset changes and remove the static route if needed.

## Open Questions

- Which DatastarUI primitives need adaptation for this app's markup, accessibility, and Datastar attributes after being copied in?
