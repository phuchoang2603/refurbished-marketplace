package metric

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Config controls the shared meter provider. Empty Endpoint records metrics
// without exporting them; the exporter itself reads OTEL_EXPORTER_OTLP_* and
// the reader reads OTEL_METRIC_EXPORT_INTERVAL.
type Config struct {
	ServiceName string
	Endpoint    string
}

// LoadConfig reads OTEL_SERVICE_NAME and the OTLP metrics endpoint.
func LoadConfig(defaultServiceName string) Config {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = strings.TrimSpace(defaultServiceName)
	}
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"))
	}
	return Config{
		ServiceName: serviceName,
		Endpoint:    endpoint,
	}
}

// Init installs the global MeterProvider used by otelhttp/otelgrpc and pushes
// its metrics over OTLP/gRPC. Shutdown flushes pending metrics.
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

	opts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if strings.TrimSpace(cfg.Endpoint) != "" {
		exp, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
	}

	mp := sdkmetric.NewMeterProvider(opts...)
	otel.SetMeterProvider(mp)
	return mp.Shutdown, nil
}
