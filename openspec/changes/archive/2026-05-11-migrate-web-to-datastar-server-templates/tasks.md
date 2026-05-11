## 1. Web Rendering Foundation

- [x] 1.1 Add `templ` component structure for the web edge pages.
- [x] 1.2 Introduce template rendering helpers in `services/web` for full-page and partial responses.
- [x] 1.3 Add shared layout and reusable fragment conventions for Datastar-compatible updates.
- [x] 1.4 Decide whether any DatastarUI primitives should be adopted as thin wrappers.

## 2. Route Conversion

- [x] 2.1 Convert public catalog and auth routes in `services/web` to return HTML instead of JSON.
- [x] 2.2 Update protected flows to render HTML pages or fragments after gRPC-backed actions succeed.
- [x] 2.3 Preserve health and webhook routes with their existing non-HTML behavior where needed.

## 3. Datastar Interaction Support

- [x] 3.1 Add Datastar-compatible fragment responses for targeted updates on key screens.
- [x] 3.2 Ensure server responses include the DOM IDs and markup structure needed for patching.
- [x] 3.3 Verify multi-step interactions can be served with successive HTML or SSE updates when required.
- [x] 3.4 Keep Datastar attributes and actions embedded in templates rather than a separate client-side layer.

## 4. Verification and Cleanup

- [x] 4.1 Update or add tests for HTML rendering, fragment responses, and auth boundary behavior.
- [x] 4.2 Remove obsolete JSON response shaping from the web edge where it is no longer used.
- [x] 4.3 Validate that internal gRPC client usage remains unchanged for browser-facing flows.
