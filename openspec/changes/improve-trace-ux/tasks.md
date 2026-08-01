## 1. Remove mesh tracing

- [x] 1.1 Delete `mesh.tracing` chart values and Gateway Telemetry / DestinationRule / istiod `otel-vt` provider (no disabled toggle left behind)
- [x] 1.2 Confirm marketplace + istiod Helm renders contain no `ecommerce-tracing`, `vtsingle-otlp-plain`, or `otel-vt`
- [x] 1.3 After deploy, verify Explore checkout TraceIds no longer include `ecommerce-ingress` / `ecommerce-waypoint` spans

## 2. Web HTTP span naming

- [x] 2.1 Update web router OpenTelemetry middleware so server spans use `METHOD` + chi route pattern and set `http.route`
- [x] 2.2 Verify parameterized routes use placeholders (e.g. `/products/{id}`) rather than raw path IDs

## 3. Shared Postgres query spans

- [x] 3.1 Add otelsql (or equivalent) dependency and wrap `*sql.DB` in `shared/runtime.OpenPostgres`
- [x] 3.2 Ensure statement attributes are truncated and do not include bound parameter values
- [x] 3.3 Run `tidy` / module sync for affected modules
- [x] 3.4 Verify a traced gRPC path shows DB child spans under the same TraceId in VictoriaTraces

## 4. Shared Redis command spans

- [x] 4.1 Enable go-redis OpenTelemetry instrumentation in `shared/runtime.OpenRedis`
- [x] 4.2 Verify a traced cart path shows Redis child spans under the same TraceId

## 5. Documentation

- [x] 5.1 Update `docs/observability.md` for app-only waterfalls (no mesh proxy span expectations)
- [x] 5.2 Document example TraceQL / Explore steps scoped to marketplace app services plus DB/Redis children
- [x] 5.3 Note that Istio L7 metrics / Marketplace Istio RED remain the mesh SLI path
