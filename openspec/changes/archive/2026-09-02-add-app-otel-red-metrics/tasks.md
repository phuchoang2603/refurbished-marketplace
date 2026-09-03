## 1. Shared metrics bootstrap

- [x] 1.1 Add `shared/observe/metric` (MeterProvider, resource `service.name`, OTLP/HTTP exporter to `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`, noop when unset, shutdown flush). Do not reuse the traces endpoint.
- [x] 1.2 Add `runtime.InitMetrics` (load config, init, log enabled/disabled) and wire `go.work` / `tidy` for the new module.

## 2. Service wiring

- [x] 2.1 Call `InitMetrics` (and defer shutdown) from `web`, `users`, `products`, `orders`, `cart`, and `payment` mains next to `InitTracing`. Skip the payment-gateway-simulator.
- [x] 2.2 Confirm `otelhttp` HTTP metrics carry chi route patterns; if series use raw URL paths, fix metric attributes on the existing web middleware only (no second histogram).

## 3. Helm

- [x] 3.1 Confirm the in-cluster VMSingle Service name and port `8428`, then set `defaults.otel.metricsEndpoint` to `http://<vmsingle>:8428/opentelemetry/v1/metrics`.
- [x] 3.2 Render `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` on marketplace Deployments that already get traces env. Leave `OTEL_EXPORTER_OTLP_ENDPOINT` pointed at VTSingle.

## 4. Grafana dashboard

- [x] 4.1 Enable `customDashboards` and add a Marketplace RED dashboard JSON (web HTTP rate/error/p95 by route; gRPC **server** rate/error/p95 by service and method). Query VictoriaMetrics only.
- [x] 4.2 After the first live export on talos-dev, align PromQL with actual metric and label names VM stored.

## 5. Docs and close-out

- [x] 5.1 Update `docs/observability.md` (and Cilium observe notes that still treat #43 as a gap) with dual OTLP destinations, dashboard location, and Hubble-not-required.
- [x] 5.2 On talos-dev: generate HTTP + gRPC traffic (browse + checkout), confirm series in VictoriaMetrics and panels on the dashboard, confirm traces still land in VictoriaTraces.
- [x] 5.3 Close GitHub issue #43 when the dashboard shows live app RED.
