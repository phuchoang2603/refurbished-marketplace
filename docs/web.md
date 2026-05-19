# Web Edge Architecture

`services/web` owns the public browser surface for the marketplace. It renders server-side HTML with `templ`, Tailwind CSS, Datastar, and vendored DatastarUI-style components, and it supports Datastar-compatible fragment updates for interactive browser flows.

## Responsibilities

- Render marketplace pages and fragments on the server.
- Compose repeated UI controls from vendored DatastarUI-style component packages.
- Keep auth/session validation at the browser boundary.
- Compose data from internal services through the existing gRPC clients.
- Preserve non-browser JSON contracts where needed, such as `GET /healthz` and the Stripe simulator webhook.

## Browser Contract

Browser routes should return HTML pages or HTML fragments. Forms should submit standard form fields to `services/web`; handlers translate those requests into gRPC calls.

Do not add browser-facing JSON DTOs unless there is a concrete external consumer. JSON belongs to explicit non-browser endpoints.

## Templates And Fragments

- Template sources live in `services/web/internal/views/*.templ`.
- Generated files live next to sources as `*_templ.go`.
- Full pages use `AppShell`.
- Auth pages and logout use native browser forms.
- Fragment responses should render elements with stable IDs that match the Datastar patch target, for example `#cart`.

## Tests

Web tests live in `services/web/tests/` and should exercise public behavior from outside the internal packages where possible.

Run the web test suite with:

```bash
cd services/web && go test ./...
```
