## Why

The current `services/web` layer exposes a JSON REST API at the browser edge, but the app is headed toward a server-rendered experience using Datastar and `templ` components. Moving the browser contract to HTML fragments and server-rendered pages removes the extra JSON mapping layer while keeping the existing gRPC service boundaries intact.

## What Changes

- Convert `services/web` from a JSON-centric REST BFF into an HTML-rendering edge service.
- Serve full pages and partial HTML responses suitable for Datastar-driven interactions.
- Use `templ` as the primary server-rendering layer, with `DatastarUI` only if it stays thin and compatible.
- Keep internal communication to `users`, `products`, `orders`, `cart`, and `payment` over gRPC.
- Remove or de-emphasize browser-facing JSON response shapes where HTML responses now own the public contract. **BREAKING**
- Preserve auth/session handling at the web edge while shifting presentation logic to server templates.

## Capabilities

### New Capabilities

- `server-rendered-web`: browser-facing HTML page and fragment rendering for the marketplace UI, built on `templ` and Datastar-compatible patterns.

### Modified Capabilities

- `web`: the public web contract changes from REST/JSON endpoints to server-rendered HTML and Datastar-compatible interactions.

## Impact

- `services/web` handlers, response shaping, and routing behavior.
- Browser clients that currently consume JSON from the web edge.
- `templ` rendering and any optional Datastar UI component dependency added to the web service.
- No change to the internal gRPC service contracts in `shared/proto/`.
