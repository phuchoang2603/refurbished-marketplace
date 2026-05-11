## Context

`services/web` currently serves the public browser edge as a JSON REST API and translates those requests into gRPC calls to internal services. The proposed change keeps the gRPC boundary intact but shifts the browser contract to server-rendered HTML built with `templ` and Datastar-friendly partial updates.

Datastar supports HTML responses that morph matching DOM elements by `id`, and it also supports SSE-based patching for incremental updates. That gives the web service two viable interaction styles for the same server-rendered edge.

`templ` is the rendering foundation, while DatastarUI is only a possible thin component kit on top. The design should keep markup explicit and componentized without depending on a second heavyweight UI system.

## Goals / Non-Goals

**Goals:**

- Replace JSON-first browser responses with HTML pages and fragments.
- Keep internal service communication over gRPC.
- Preserve the web service as the public auth and composition boundary.
- Enable interactive browser updates without introducing a client-side SPA state layer.
- Use `templ` for typed, reusable server-side components and keep `DatastarUI` optional.

**Non-Goals:**

- Changing the internal protobuf contracts or service boundaries.
- Introducing a new browser RPC stack such as gRPC-Web.
- Reworking service persistence or eventing.
- Designing the full page IA or visual system.

## Decisions

- Use server-rendered templates as the primary browser contract.
  - Rationale: the existing `web` service already sits at the edge and aggregates multiple gRPC services; templates let it own composition and auth while avoiding JSON DTO duplication.
  - Alternatives considered: keep REST/JSON and add a frontend client generator; expose gRPC-Web directly; use raw `html/template`. The first two preserve the mapping layer this change is trying to remove, while `html/template` gives up typed reusable components that `templ` provides.

- Use `templ` as the component system and keep DatastarUI optional.
  - Rationale: `templ` gives compile-time safety, reusable fragments, and a clean fit with Datastar's HTML-first model. DatastarUI can accelerate UI assembly later if it remains thin and `templ`-native.
  - Alternatives considered: build on raw `html/template` only; adopt a large external component library as the primary UI system. Both increase either fragility or coupling compared to a `templ`-first approach.

- Prefer HTML fragment responses for common interactions, with SSE-style patching reserved for progressive or multi-step updates.
  - Rationale: Datastar can morph HTML responses by element id, which keeps simple interactions close to standard HTTP while still supporting reactive updates. SSE is useful when the server needs to stream successive states.
  - Alternatives considered: a pure SPA-style hydration model; a JSON API consumed by Datastar. Both add extra client logic compared to server-owned HTML.

- Keep `services/web` as the single browser entrypoint.
  - Rationale: auth/session checks, request normalization, and multi-service composition already live here; moving templates into the same service keeps edge concerns together.
  - Alternatives considered: split rendering into a separate frontend service. That would duplicate edge concerns and complicate deployment for little gain at this stage.

- Treat HTML fragments as the stable browser contract and deprecate JSON response shapes only where the browser previously depended on them.
  - Rationale: the browser should consume markup, not a parallel JSON schema, once Datastar is the interaction model.
  - Alternatives considered: keep both HTML and JSON forever. That would extend maintenance cost and blur ownership of the browser contract.

## Risks / Trade-offs

- [Risk] Template responses can entangle presentation with edge logic more tightly than JSON handlers.
  - Mitigation: keep gRPC client calls and data shaping in handlers, while moving markup into template components/files.

- [Risk] Datastar fragment updates may be awkward for flows that expect full-page navigation.
  - Mitigation: use normal HTML responses for page loads and reserve Datastar interactions for targeted updates.

- [Risk] This is a breaking change for any browser code consuming JSON from `services/web`.
  - Mitigation: scope the change to the web-facing contract and document the affected routes clearly in the specs.

- [Risk] Template and fragment conventions may drift across screens without an agreed structure.
  - Mitigation: define a small set of page and fragment patterns before implementation begins.

- [Risk] A UI kit can become a second design system if adopted too early.
  - Mitigation: keep DatastarUI optional and limit it to reusable primitives only if it matches the `templ` component model.
