package runtime

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	sharedmetric "github.com/phuchoang2603/refurbished-marketplace/shared/observe/metric"
)

// InitMetrics configures the global OpenTelemetry meter provider.
// Empty OTEL_EXPORTER_OTLP_METRICS_ENDPOINT keeps a noop exporter.
func InitMetrics(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	cfg := sharedmetric.LoadConfig(serviceName)
	shutdown, err := sharedmetric.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Endpoint == "" {
		sharedlog.Info("metrics disabled (set OTEL_EXPORTER_OTLP_METRICS_ENDPOINT to enable)")
	} else {
		sharedlog.Info("metrics enabled", "endpoint", cfg.Endpoint)
	}
	return shutdown, nil
}
