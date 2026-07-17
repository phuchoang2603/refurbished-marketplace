## 1. Platform scrapes and OTLP destination

- [x] 1.1 Enable VTSingle OTLP gRPC (`-otlpGRPCListenAddr=:4317`, `-otlpGRPC.tls=false`) and expose Service port 4317; document HTTP fallback `http://vtsingle-vmks.monitoring.svc:10428/insert/opentelemetry/v1/traces`.
- [x] 1.2 Update `istioScrapes` to keep waypoint + add ingress Gateway scrapes; remove istiod, ztunnel, and cni scrapes.
- [x] 1.3 Update `docs/observability.md` scrape table to match waypoint + ingress only; document direct OTLP (no collector).

## 2. Shared Go tracing bootstrap

- [x] 2.1 Add shared OpenTelemetry bootstrap (provider, W3C propagators, resource attrs, OTLP exporter) under `shared/`.
- [x] 2.2 Wire bootstrap into service runtimes via env (endpoint, service name, sampling) without requiring Tilt to deploy VT.

## 3. Sync path instrumentation

- [x] 3.1 Export web spans and inject W3C context on outgoing gRPC clients (checkout + hosted-payment callback paths).
- [x] 3.2 Add `otelgrpc` (or equivalent) server/client instrumentation on orders, payment, and other services on the checkout/callback paths.
- [x] 3.3 Verify a sync-only TraceId appears in Grafana Explore for web → gRPC before enabling outbox work.

## 4. Outbox span context

- [x] 4.1 Add `tracingspancontext` migrations for `orders_outbox`, `payment_outbox`, and `inventory_outbox`.
- [x] 4.2 Update sqlc queries and outbox writers to persist active span context in the same transaction.
- [x] 4.3 Update Kafka consumers to extract `traceparent` and create child spans.

## 5. Debezium / Connect tracing

- [x] 5.1 Rely on Strimzi Kafka image OpenTelemetry jars + enable Connect `tracing.type: opentelemetry` (agent); keep Debezium plugin in `connect-debezium` image.
- [x] 5.2 Configure EventRouter tracing fields on orders/payment/inventory outbox connectors.
- [ ] 5.3 Rebuild Connect image if needed, sync Kafka chart, and confirm connector tasks emit Kafka `traceparent`.

## 6. Istio mesh tracing

- [x] 6.1 Configure Istio OpenTelemetry extension provider toward VictoriaTraces.
- [x] 6.2 Apply Telemetry resources for ecommerce ingress and waypoint with staging-friendly sampling.
- [ ] 6.3 Verify mesh spans share the app TraceId when headers are propagated.

## 7. Verification and close-out

- [ ] 7.1 Exercise staging checkout and confirm one connected TraceId through outbox → Debezium → inventory.
- [ ] 7.2 Exercise hosted-payment success/fail callback TraceId through payment path.
- [x] 7.3 Document e2e tracing architecture, joining rules, and troubleshooting in `docs/observability.md`.
- [x] 7.4 Link acceptance to GitHub issue #3 and run OpenSpec validation for this change.
