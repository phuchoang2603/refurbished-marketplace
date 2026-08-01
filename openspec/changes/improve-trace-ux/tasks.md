## 1. Disable mesh tracing

- [ ] 1.1 Set `mesh.tracing.enabled: false` in marketplace chart values (local + staging overlay as needed)
- [ ] 1.2 Confirm `templates/tracing.tpl` no longer renders `Telemetry` `ecommerce-tracing` when disabled
- [ ] 1.3 After deploy, verify Explore checkout TraceIds no longer include `ecommerce-ingress` / `ecommerce-waypoint` spans

## 2. Web HTTP span naming

- [ ] 2.1 Update web router OpenTelemetry middleware so server spans use `METHOD` + chi route pattern and set `http.route`
- [ ] 2.2 Verify parameterized routes use placeholders (e.g. `/products/{id}`) rather than raw path IDs

## 3. Shared Postgres query spans

- [ ] 3.1 Add otelsql (or equivalent) dependency and wrap `*sql.DB` in `shared/runtime.OpenPostgres`
- [ ] 3.2 Ensure statement attributes are truncated and do not include bound parameter values
- [ ] 3.3 Run `tidy` / module sync for affected modules
- [ ] 3.4 Verify a traced gRPC path shows DB child spans under the same TraceId in VictoriaTraces

## 4. Shared Redis command spans

- [ ] 4.1 Enable go-redis OpenTelemetry instrumentation in `shared/runtime.OpenRedis`
- [ ] 4.2 Verify a traced cart path shows Redis child spans under the same TraceId

## 5. Documentation

- [ ] 5.1 Update `docs/observability.md` for app-only waterfalls (no mesh proxy span expectations)
- [ ] 5.2 Document example TraceQL / Explore steps scoped to marketplace app services plus DB/Redis children
- [ ] 5.3 Note that Istio L7 metrics / Marketplace Istio RED remain the mesh SLI path
