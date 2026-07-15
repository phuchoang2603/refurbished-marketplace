## Context

Staging already runs VictoriaTraces (VTSingle) with a Grafana Tempo-compatible datasource (`/select/tempo`, TraceQL/Explore). Istio ambient enrolls marketplace workloads with a waypoint and an ingress Gateway. Metrics scrapes today cover istiod, ztunnel, cni, and waypoint via `VMPodScrape`; ingress Envoy is not scraped.

Application tracing is incomplete: `services/web` applies `otelhttp` without a configured exporter or gRPC propagation. Outbox tables lack span context. Debezium `EventRouter` connectors do not map tracing fields. Kafka consumers do not continue traces. Issue #3 defines the desired HTTP → gRPC → outbox → Debezium → Kafka → consumer path.

## Goals / Non-Goals

**Goals:**

- One W3C **TraceId** for a checkout (and hosted-payment callback) spanning app spans, Istio L7 hops, Debezium, and consumers.
- Export app and mesh spans to VictoriaTraces; visualize in Grafana Explore.
- Persist span context on outbox rows; EventRouter emits `traceparent` Kafka headers.
- Child-of async span relation for Grafana waterfall UX.
- OpenTelemetry SDK on the `connect-debezium` image.
- Metrics scrapes limited to **waypoint + ingress**.

**Non-Goals:**

- OTEL messaging **links** instead of child-of (v1).
- Enrolling Kafka Connect into ambient mesh solely for tracing.
- Per-service custom `/metrics` or log↔trace correlation (#2).
- Production-grade sampling/tail sampling beyond simple staging defaults.
- Replacing Debezium outbox with a polling publisher.

## Decisions

### 1. Shared Go OTEL bootstrap exports OTLP to VictoriaTraces

Introduce `shared/observability/trace` (or equivalent module path) that configures a TracerProvider, W3C `TraceContext` + `Baggage` propagators, resource attributes (`service.name`), and OTLP HTTP/gRPC exporter aimed at VTSingle (or a thin collector in front if required by chart defaults).

**Rationale:** One bootstrap avoids divergent exporters per service; VT already accepts OTLP/Jaeger-compatible ingest paths used by the stack.

**Alternatives considered:** Jaeger agent sidecars (extra pods); stdout-only traces (not useful in Grafana).

### 2. Sync path: app OTEL + Istio mesh tracing, same TraceId via W3C headers

- Apps extract/inject `traceparent` / `tracestate` (and forward `x-request-id` when present).
- gRPC clients/servers use `otelgrpc`.
- Istio `meshConfig.extensionProviders` OpenTelemetry provider points at VictoriaTraces (or collector); `Telemetry` resources enable tracing on waypoint and ingress Gateways with W3C propagation.

**Rationale:** Istio docs require apps to propagate headers for proxy spans to join the same TraceId. Ambient distributed tracing is waypoint/ingress L7, not ztunnel.

**Alternatives considered:** Mesh-only tracing (misses business/DB/outbox spans); app-only tracing (misses hop timing at the proxy).

### 3. Async path: outbox column + EventRouter tracing + child-of consumers

- Add `tracingspancontext` (text) to `orders_outbox`, `payment_outbox`, `inventory_outbox`.
- On outbox insert, serialize the active span context (W3C propagator → Properties/string) in the same transaction.
- Configure Debezium `EventRouter` tracing (`tracing.span.context.field`, etc.) so Connect restores context, emits a span, and sets Kafka `traceparent`.
- Consumers extract headers and start **child** spans (parent–child), not links, for v1.

**Rationale:** Official Debezium outbox tracing model; CDC process cannot see the writer’s in-memory context without a DB carrier. Child-of matches issue #3 / tutorial UX in Grafana.

**Alternatives considered:** OTEL span links (better semantics for long lag, weaker Explore UX); custom SMT instead of EventRouter (we already use EventRouter); skipping Connect OTEL (headers alone without Connect spans is weaker).

### 4. Add OpenTelemetry SDK to `connect-debezium` image

Bake OTEL API/SDK (and any Debezium-required tracing deps) into `infra/docker/connect-debezium.Dockerfile` / Connect runtime so EventRouter tracing can activate.

**Rationale:** Debezium tracing requires OTEL on the Connect classpath; Strimzi KafkaConnect image alone is insufficient.

**Alternatives considered:** Init container copying jars (fragile); separate collector-only story without Connect spans (loses Debezium span in the waterfall).

### 5. Slim Istio scrapes to waypoint + ingress

Remove `VMPodScrape` for istiod, ztunnel, and cni. Keep waypoint; add ingress Gateway scrape (`ecommerce-ingress` / Envoy prometheus ports).

**Rationale:** App/edge SLIs come from L7 proxies; control-plane scrapes are platform noise for this marketplace.

**Alternatives considered:** Keep all ambient scrapes (noise); scrape only waypoint (misses north-south edge).

### 6. Staging-friendly sampling first

Use high or 100% sampling in staging Telemetry / SDK defaults; document that production will tighten later.

**Rationale:** Verify joined TraceIds before optimizing cost.

## Risks / Trade-offs

| Risk                                                              | Mitigation                                                                |
| ----------------------------------------------------------------- | ------------------------------------------------------------------------- |
| Apps forget to propagate headers → split TraceIds                 | Shared gRPC dial/server helpers; acceptance checks for one TraceId        |
| Connect OTEL jars / version skew with Debezium                    | Pin versions in Dockerfile; verify connector task starts                  |
| Child-of makes slow consumers inflate perceived checkout duration | Accept for v1; revisit links if lag becomes common                        |
| VT OTLP endpoint / auth mismatch                                  | Document exact URL/port from chart; smoke-test with one service first     |
| Ingress scrape port name differs from waypoint                    | Confirm Service/pod ports on `ecommerce-ingress-istio` before scrape CR   |
| Outbox migration + connector config must roll together            | Order: migrate DBs → ship Connect image → update connector tracing fields |

## Migration Plan

1. Deploy scrape diet + confirm waypoint/ingress metrics still useful.
2. Ship shared OTEL bootstrap; enable export from `web` + one gRPC service; confirm spans in Grafana.
3. Add gRPC instrumentation across orders/payment/products/cart/users as needed for checkout paths.
4. Outbox migrations + writers; rebuild Connect with OTEL; enable EventRouter tracing; update consumers.
5. Enable Istio Telemetry → VT; verify mesh spans share TraceId with app spans.
6. Document verification (checkout + hosted-payment callback) in `docs/observability.md`.
7. Rollback: disable Telemetry CRs / unset OTEL endpoint env; scrapes can revert via Git; outbox column is additive (nullable/unused if writers stop).

## Resolved questions

### VTSingle OTLP address (cluster inspected 2026-07-14)

Live Service: `vtsingle-vmks.monitoring.svc.cluster.local:10428` (only port today).

- Pod listens on **10428 only**; OTLP **gRPC is disabled** until `-otlpGRPCListenAddr` is set.
- **HTTP OTLP (works now):**  
  `http://vtsingle-vmks.monitoring.svc.cluster.local:10428/insert/opentelemetry/v1/traces`  
  Confirmed path exists on the running VTSingle (`v0.9.4`).
- **gRPC OTLP (preferred if we enable it):** enable via VTSingle `extraArgs`, e.g.  
  `-otlpGRPCListenAddr=:4317` and `-otlpGRPC.tls=false` for in-cluster, then expose Service port **4317**.  
  Go SDK: `otlptracegrpc` → `vtsingle-vmks.monitoring.svc:4317` with insecure TLS for staging.

**Decision:** Prefer **direct OTLP/gRPC** after enabling `:4317` on VTSingle + Service; HTTP path is the zero-config fallback. No OpenTelemetry Collector required for v1 (apps and Istio export straight to VT).

### Consumers to instrument

Instrument **all** marketplace Kafka consumers (not only #3’s minimum):

| Service                | Topics consumed                                                       |
| ---------------------- | --------------------------------------------------------------------- |
| `products` (inventory) | `orders.created`, `payment.succeeded`, `payment.failed`               |
| `payment`              | `inventory.reserved`                                                  |
| `orders`               | `inventory.reservation-failed`, `payment.succeeded`, `payment.failed` |

Also instrument writers/outboxes on those paths so TraceIds continue through Debezium for every outbox entity.
