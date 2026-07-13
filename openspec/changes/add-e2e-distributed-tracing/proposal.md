## Why

VictoriaTraces and Grafana Explore already work, but checkout flows still lack connected distributed traces: `web` only wraps `otelhttp` without exporting or propagating context, outbox rows have no span context, Debezium does not emit `traceparent` on Kafka headers, and consumers do not continue traces. Issue [#3](https://github.com/phuchoang2603/refurbished-marketplace/issues/3) asks for end-to-end tracing across HTTP → gRPC → transactional outbox → Debezium → Kafka → consumers, with Istio mesh hops sharing the same W3C TraceId.

## What Changes

- Add a shared Go OpenTelemetry package (tracer provider, W3C propagators, OTLP export to VictoriaTraces) and wire it into marketplace services.
- Propagate `traceparent` on web → gRPC clients and instrument gRPC servers (`otelgrpc`).
- Persist span context on outbox writes (`tracingspancontext` column) in the same DB transaction as domain events for orders, payment, and inventory outboxes.
- Add the OpenTelemetry SDK to the `connect-debezium` image and configure Debezium `EventRouter` tracing so Kafka records carry `traceparent`.
- Have Kafka consumers extract context and create **child-of** spans (same TraceId; parent–child async relation for Grafana waterfall UX).
- Enable Istio waypoint (and ingress) OpenTelemetry tracing to VictoriaTraces so mesh hops join the same TraceId when apps forward W3C headers.
- Slim Istio metrics scrapes to **waypoint + ingress** only; remove istiod / ztunnel / cni `VMPodScrape` targets.
- Document the e2e path, TraceId joining rules, and verification in `docs/observability.md`.

## Capabilities

### New Capabilities

- `distributed-tracing`: End-to-end W3C tracing across HTTP/gRPC, outbox persistence, Debezium/Kafka headers, and consumers, exporting to VictoriaTraces for Grafana visualization.

### Modified Capabilities

- `platform-observability`: Narrow Istio scrape targets to waypoint + ingress; acknowledge application/mesh OTLP emission into VictoriaTraces (no longer metrics-only first slice for traces).
- `istio-observability`: Require mesh tracing export (OpenTelemetry / W3C) for enrolled L7 hops so proxy spans share app TraceIds when headers are propagated.
- `web`: Extend web-edge tracing beyond in-process `otelhttp` to export spans and inject context on outgoing gRPC calls (including hosted-payment callback path).

## Impact

- Touches `shared/` (new observability/trace helpers), `services/web`, `services/orders`, `services/payment`, `services/products` (gRPC + outbox migrations/sqlc), Kafka consumers, `infra/docker/connect-debezium.Dockerfile`, `infra/charts/kafka` connector config, `infra/charts/observability` scrapes, Istio `meshConfig` / `Telemetry` for OTLP to VT.
- Depends on existing VTSingle + Grafana Jaeger datasource (already verified).
- Does not change business protobuf payloads or browser UX contracts.
- Non-goals: full auto-instrumentation of every library call; OTEL messaging **links** instead of child-of for v1; production sampling policy beyond staging-friendly defaults; enrolling Kafka Connect into ambient mesh solely for tracing.
