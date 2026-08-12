package runtime

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
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
		sharedlog.Info("tracing disabled (set OTEL_EXPORTER_OTLP_ENDPOINT to enable)")
	} else {
		sharedlog.Info("tracing enabled", "endpoint", cfg.Endpoint)
	}
	return shutdown, nil
}
