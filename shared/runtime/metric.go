package runtime

import (
	"context"
	"errors"
	"net/http"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	sharedmetric "github.com/phuchoang2603/refurbished-marketplace/shared/observe/metric"
)

// InitMetrics configures the global OpenTelemetry meter provider for Prometheus
// scrape. METRICS_ADDR=- skips the /metrics listener (process still starts).
func InitMetrics(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	cfg := sharedmetric.LoadConfig(serviceName)
	shutdownMP, err := sharedmetric.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Addr == "" {
		sharedlog.Info("metrics scrape listener disabled (METRICS_ADDR=-)")
		return shutdownMP, nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", sharedmetric.Handler())
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		sharedlog.Info("metrics scrape enabled", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sharedlog.Error("metrics listener", "err", err)
		}
	}()

	return func(ctx context.Context) error {
		shutdownErr := srv.Shutdown(ctx)
		return errors.Join(shutdownErr, shutdownMP(ctx))
	}, nil
}
