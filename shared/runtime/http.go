package runtime

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const defaultHTTPShutdownTimeout = 30 * time.Second

type HTTPServerConfig struct {
	Addr            string
	ServiceName     string
	Handler         http.Handler
	ShutdownTimeout time.Duration
}

func ServeHTTP(ctx context.Context, cfg HTTPServerConfig) error {
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: cfg.Handler,
	}

	shutdownTimeout := cfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultHTTPShutdownTimeout
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting http service", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	return nil
}
