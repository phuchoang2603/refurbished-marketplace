package runtime

import (
	"context"
	"log"

	sharedtrace "refurbished-marketplace/shared/trace"
)

// InitTracing configures the global OpenTelemetry provider.
// Empty OTEL_EXPORTER_OTLP_ENDPOINT keeps a noop exporter (Tilt-friendly).
func InitTracing(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	cfg := sharedtrace.LoadConfig(serviceName)
	shutdown, err := sharedtrace.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Endpoint == "" {
		log.Printf("tracing: disabled (set OTEL_EXPORTER_OTLP_ENDPOINT to enable)")
	} else {
		log.Printf("tracing: exporting from %s to %s", cfg.ServiceName, cfg.Endpoint)
	}
	return shutdown, nil
}
