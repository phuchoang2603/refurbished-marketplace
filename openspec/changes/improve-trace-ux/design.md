## Context

E2e distributed tracing already joins one W3C TraceId across HTTP → gRPC → outbox → Debezium → Kafka → consumers into VictoriaTraces. Istio ambient Telemetry also exports ingress/waypoint spans, so Explore often roots on `ecommerce-ingress` / `ecommerce-waypoint`. Web uses `otelhttp.NewMiddleware("web")`, so the app server span name is literally `web`. Postgres opens via raw `database/sql` with no driver tracer; Redis via plain `go-redis` with no OTEL hook. gRPC (`otelgrpc`) and Kafka consumer spans already provide per-RPC / per-message entry spans.

## Goals / Non-Goals

**Goals:**

- Operation-centric Explore UX: readable HTTP / RPC / messaging parents with DB and Redis children.
- Disable marketplace Istio mesh tracing (no proxy spans in the default waterfall).
- Keep ambient mesh + Istio L7 metrics / RED dashboard.
- Instrument once at shared boundaries (`OpenPostgres`, `OpenRedis`, web middleware) so every gRPC method and Kafka handler gets coverage without per-method `span.Start`.

**Non-Goals:**

- Hand-rolled domain spans on every service method.
- Named sqlc-only wrappers instead of driver-level query spans.
- Canary traffic splitting or version-based promotion (later; needs metrics + version attrs, not mesh traces).
- Changing sampling policy beyond existing staging defaults.
- Requiring outbound HTTP client instrumentation before any service uses it (add when needed).

## Decisions

### 1. Turn off Istio mesh tracing (not just filter it)

Set `mesh.tracing.enabled: false` (local + staging overlays as needed) so `templates/tracing.tpl` does not render `Telemetry` `ecommerce-tracing`. Leave istiod `extensionProviders` / `otel-vt` in place if harmless; unused when no Telemetry attaches.

**Rationale:** Proxy spans confuse root naming and are unnecessary for app+DB deep-dives. Canary later relies on versioned metrics, not Envoy spans. Reversible by flipping the flag.

**Alternatives considered:** Keep exporting and demote via TraceQL only (still clutters service pickers); remove extension provider entirely (more churn, little gain).

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
| Losing proxy-only latency visibility                         | Keep Istio RED metrics; re-enable Telemetry if Envoy debugging is needed                    |
| Chi route pattern empty if middleware order wrong            | Set/update span name after match (`chi.RouteContext` / otelchi-style)                       |
| Test DBs also wrapped                                        | Prefer wrapping only production opener path; tests can stay uninstrumented or noop exporter |

## Migration Plan

1. Disable mesh tracing in chart values; sync local/staging; confirm no new ingress/waypoint spans.
2. Ship web route-pattern naming; confirm Explore shows `METHOD /route`.
3. Add otelsql to `OpenPostgres` + module tidy; verify SQL children under CreateOrder / consumers.
4. Add Redis OTEL hook; verify cart path.
5. Update docs and remove mesh-from-waterfall expectations.
6. Rollback: re-enable `mesh.tracing.enabled`; revert opener wrappers / middleware (deps stay unused).

## Open Questions

- None blocking; pick concrete otelsql / redisotel module versions at implement time via go modules / Context7 if needed.
