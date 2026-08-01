## Context

E2e distributed tracing already joins one W3C TraceId across HTTP → gRPC → outbox → Debezium → Kafka → consumers into VictoriaTraces. Istio ambient Telemetry also exports ingress/waypoint spans, so Explore often roots on `ecommerce-ingress` / `ecommerce-waypoint`. Web uses `otelhttp.NewMiddleware("web")`, so the app server span name is literally `web`. Postgres opens via raw `database/sql` with no driver tracer; Redis via plain `go-redis` with no OTEL hook. gRPC (`otelgrpc`) and Kafka consumer spans already provide per-RPC / per-message entry spans.

## Goals / Non-Goals

**Goals:**

- Operation-centric Explore UX: readable HTTP / RPC / messaging parents with DB and Redis children.
- Remove marketplace Istio mesh tracing (no proxy spans; no leftover toggle).
- Keep ambient mesh + Istio L7 metrics / RED dashboard.
- Instrument once at shared boundaries (`OpenPostgres`, `OpenRedis`, web middleware) so every gRPC method and Kafka handler gets coverage without per-method `span.Start`.

**Non-Goals:**

- Hand-rolled domain spans on every service method.
- Named sqlc-only wrappers instead of driver-level query spans.
- Canary traffic splitting or version-based promotion (later; needs metrics + version attrs, not mesh traces).
- Changing sampling policy beyond existing staging defaults.
- Requiring outbound HTTP client instrumentation before any service uses it (add when needed).

## Decisions

### 1. Remove Istio mesh tracing (no toggle)

Delete marketplace Gateway Telemetry (`templates/tracing.tpl`), `mesh.tracing` chart values, the `vtsingle-otlp-plain` DestinationRule (mesh.tpl), and the istiod `otel-vt` `extensionProviders` entry. Do not leave an `enabled: false` flag.

**Rationale:** Proxy spans confuse root naming and are unnecessary for app+DB deep-dives. Canary later relies on versioned metrics, not Envoy spans. Dead config is worse than a clean cut.

**Alternatives considered:** Keep exporting and demote via TraceQL only (still clutters service pickers); feature-flag off but keep Telemetry/provider templates (stale path).

### 2. HTTP span names = `METHOD` + chi route pattern

Replace opaque `otelhttp` operation-as-name with a formatter (or chi-aware middleware) that sets the span name to e.g. `POST /orders/{id}` after route match, and sets `http.route` to the pattern. Keep `service.name=web`.

**Rationale:** Cardinality-safe, matches OTEL HTTP semantic conventions, readable in Explore lists.

**Alternatives considered:** Raw URL path (ID cardinality explosion); keep `"web"` (status quo).

### 3. All Postgres queries via otelsql in `OpenPostgres`

Wrap the `*sql.DB` returned by `shared/runtime.OpenPostgres` with otelsql (or equivalent `database/sql` driver wrapper). sqlc and `BeginTx` inherit context → child spans under gRPC/Kafka parents. Truncate/sanitize statement attributes; do not attach bound parameter values.

**Rationale:** One shared change covers all services; matches “all queries” preference with minimal service code.

**Alternatives considered:** Wrap each sqlc `Queries` method (better names, more code); leave DB untraced.

### 4. Redis via go-redis OTEL hook in `OpenRedis`

Enable redisotel (or current go-redis/v9 tracing hook) when creating the cart Redis client in `shared/runtime.OpenRedis`.

**Rationale:** Same “instrument at opener” pattern as Postgres.

**Alternatives considered:** Manual spans around `loadCart` / `saveCart` only.

### 5. No extra business spans on gRPC/Kafka handlers

Treat existing `otelgrpc` server spans and `messaging process <topic>` consumer spans as the entry parents. DB/Redis children nest under them via context.

**Rationale:** “Every method + handler, streamlined” without duplicate `CreateOrder` layers.

**Alternatives considered:** Explicit `trace.Start` in every service method (noise + maintenance).

### 6. Docs: app-only verification story

Update `docs/observability.md` so checkout verification expects app + Connect spans, not `ecommerce-ingress` / `ecommerce-waypoint`. Document example TraceQL scoped to marketplace app services.

## Risks / Trade-offs

| Risk                                                         | Mitigation                                                                                  |
| ------------------------------------------------------------ | ------------------------------------------------------------------------------------------- |
| otelsql span names are generic (`query` / statement keyword) | Rely on gRPC/Kafka/HTTP parents for readability; put SQL in attributes                      |
| High span volume with 100% sampling + all SQL                | Accept in staging; truncate statements; revisit sampling before prod                        |
| Losing proxy-only latency visibility                         | Keep Istio RED metrics; restore mesh tracing only as a new change if needed                 |
| Chi route pattern empty if middleware order wrong            | Set/update span name after match (`chi.RouteContext` / otelchi-style)                       |
| Test DBs also wrapped                                        | Prefer wrapping only production opener path; tests can stay uninstrumented or noop exporter |

## Migration Plan

1. Remove mesh tracing templates/values/provider; sync local/staging; confirm no ingress/waypoint spans and no leftover CRs.
2. Ship web route-pattern naming; confirm Explore shows `METHOD /route`.
3. Add otelsql to `OpenPostgres` + module tidy; verify SQL children under CreateOrder / consumers.
4. Add Redis OTEL hook; verify cart path.
5. Update docs for app-only waterfalls.
6. Rollback of app instrumentation: revert opener wrappers / middleware (deps stay unused). Mesh tracing stay removed unless a follow-up change reintroduces it.

## Open Questions

- None blocking; pick concrete otelsql / redisotel module versions at implement time via go modules / Context7 if needed.
