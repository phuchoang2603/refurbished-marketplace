package metric

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

const defaultMetricsAddr = ":9100"

// Config controls the shared meter provider and optional scrape listener.
type Config struct {
	ServiceName string
	Addr        string
}

// LoadConfig reads OTEL_SERVICE_NAME and METRICS_ADDR.
// Empty METRICS_ADDR uses :9100. Set METRICS_ADDR=- to skip the scrape listener.
func LoadConfig(defaultServiceName string) Config {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = strings.TrimSpace(defaultServiceName)
	}
	addr := strings.TrimSpace(os.Getenv("METRICS_ADDR"))
	if addr == "" {
		addr = defaultMetricsAddr
	}
	if addr == "-" {
		addr = ""
	}
	return Config{
		ServiceName: serviceName,
		Addr:        addr,
	}
}

// Init installs a Prometheus-backed global MeterProvider so otelhttp/otelgrpc
// can be scraped at /metrics. It does not push OTLP metrics.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return func(context.Context) error { return nil }, fmt.Errorf("metric: service name is required")
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	exp, err := prometheus.New(
		prometheus.WithoutScopeInfo(),
		prometheus.WithResourceAsConstantLabels(attribute.NewAllowKeysFilter(semconv.ServiceNameKey)),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exp),
	)
	otel.SetMeterProvider(mp)
	return mp.Shutdown, nil
}

// Handler serves Prometheus text for the default gatherer used by Init.
func Handler() http.Handler {
	return promhttp.Handler()
}
