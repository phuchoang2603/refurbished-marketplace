## Why

Checkout TraceIds join correctly across web → gRPC → outbox → Kafka → consumers, but Explore UX is poor: root/operation names look like infrastructure (`ecommerce-ingress`, `ecommerce-waypoint`, literal `web`), and there are no Postgres or Redis spans to deep-dive a slow request. Mesh proxy spans add noise without helping day-to-day app debugging (and are not required for a later canary, which needs versioned metrics + app attrs).

## What Changes

- **Remove Istio mesh tracing** (no Gateway Telemetry, no istiod `otel-vt` provider, no mesh-tracing DestinationRule). Keep ambient mesh and Istio L7 **metrics** scrapes / RED dashboard.
- **Name web HTTP spans** as `METHOD` + chi **route pattern** (e.g. `POST /checkout`), with `http.route` set — never raw paths with IDs.
- **Instrument all Postgres queries** via shared `OpenPostgres` (otelsql or equivalent) so sqlc calls emit child spans under existing gRPC/Kafka parents.
- **Instrument Redis** via shared `OpenRedis` (go-redis OTEL hook) for cart paths.
- **Rely on existing entry spans** (`otelgrpc`, Kafka `messaging process …`) for every RPC and consumer — no per-method hand-rolled business spans.
- **Update observability docs** for app-only waterfalls and verification without mesh span expectations.
- Optional later: outbound HTTP client helper if/when services add external HTTP (not blocking).

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `distributed-tracing`: Operation-centric span naming; shared DB/Redis instrumentation; docs and e2e expectations without mesh proxy spans as part of the default TraceId story.
- `istio-observability`: Mesh tracing to VictoriaTraces is **removed**; metrics-oriented mesh telemetry remains.
- `web`: HTTP server spans use route-pattern names (not the otelhttp operation string / service name).

## Impact

- **Code / charts:** Delete `tracing.tpl`, mesh-tracing DestinationRule, `mesh.tracing` values, istiod `otel-vt` provider; web route-span middleware; `OpenPostgres` / `OpenRedis` OTEL wrappers (`otelsql`, `redisotel`).
- **Ops:** Grafana Explore is app + Connect only; Istio RED metrics unchanged. No disabled toggle left for mesh tracing.
- **Non-goals:** Per-method domain spans; named sqlc wrappers instead of driver-level SQL spans; canary traffic splitting; restoring mesh proxy traces; production sampling policy beyond current staging defaults.
