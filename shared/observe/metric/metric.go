package metric

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Config controls the shared meter provider. Empty Endpoint skips export
// (noop provider).
type Config struct {
	ServiceName string
	Endpoint    string
}

// LoadConfig reads OTEL_EXPORTER_OTLP_METRICS_ENDPOINT and OTEL_SERVICE_NAME.
// It does not fall back to the traces OTLP endpoint.
func LoadConfig(defaultServiceName string) Config {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = strings.TrimSpace(defaultServiceName)
	}
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"))
	return Config{
		ServiceName: serviceName,
		Endpoint:    endpoint,
	}
}

// Init installs the global MeterProvider. Returns a shutdown func.
// If Endpoint is empty, installs a noop provider.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return func(context.Context) error { return nil }, fmt.Errorf("metric: service name is required")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		mp := sdkmetric.NewMeterProvider()
		otel.SetMeterProvider(mp)
		return mp.Shutdown, nil
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

	endpoint := cfg.Endpoint
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + strings.TrimPrefix(endpoint, "/")
	}
	exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
	)
	otel.SetMeterProvider(mp)
	return mp.Shutdown, nil
}
